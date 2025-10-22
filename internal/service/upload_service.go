package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/histopathai/models"
)

type Collection string

const (
	ImageCollection     Collection = "images"
	PatientCollection   Collection = "patients"
	WorkspaceCollection Collection = "workspaces"
	UserCollection      Collection = "users"
)

type UploadService struct {
	storageClient *storage.Client
	bucketName    string
	repo          *repository.MainRepository
	logger        *slog.Logger
}

type SignedUrlResponse struct {
	ImageID   string `json:"image_id"`
	UploadURL string `json:"upload_url"`
	ExpiresAt int64  `json:"expires_at"` // unix timestamp
}

type ImageInfo struct {
	FileName  string `json:"file_name"`
	Format    string `json:"format"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	SizeBytes int64  `json:"size_bytes"`
}

type PatientInfo struct {
	PatientID string  `json:"id,omitempty"` // optional, will be created if empty
	Age       *int    `json:"age,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	Race      *string `json:"race,omitempty"`
	Disease   *string `json:"disease,omitempty"`
	History   *string `json:"history,omitempty"`
}

type UploadImageRequest struct {
	ImageInfo   ImageInfo   `json:"image_info"`
	PatientInfo PatientInfo `json:"patient_info"`
	WorkspaceID string      `json:"workspace_id"`
}

func MimeTypeFromFormat(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "tiff", "tif":
		return "image/tiff"
	case "bmp":
		return "image/bmp"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func NewUploadService(
	storageClient *storage.Client,
	bucketName string,
	repo *repository.MainRepository,
	logger *slog.Logger,
) *UploadService {
	return &UploadService{
		storageClient: storageClient,
		bucketName:    bucketName,
		repo:          repo,
		logger:        logger,
	}
}

func (us *UploadService) validateUploadRequest(ctx context.Context, req *UploadImageRequest) error {
	details := make(map[string]interface{})
	// Check workspace exists
	result, err := us.repo.Read(ctx, string(WorkspaceCollection), req.WorkspaceID)
	if err != nil {
		return apperrors.NewInternalError("failed to read workspace during upload validation", err)
	}
	if result == nil {
		details["workspace"] = "required"
	} else {
		w := &models.Workspace{}
		w.FromMap(result)
		if w.AnnotationTypeID == "" {
			details["workspace"] = "workspace must have an annotation type assigned"
		}
	}

	// Check image info
	if req.ImageInfo.FileName == "" {
		details["imageInfo.file_name"] = "required"
	}
	if req.ImageInfo.Format == "" {
		details["imageInfo.format"] = "required"
	} else if !models.IsImageFormatSupported(req.ImageInfo.Format) {
		details["imageInfo.format"] = fmt.Sprintf("unsupported format: %s", req.ImageInfo.Format)
	}
	if req.ImageInfo.Width <= 0 {
		details["imageInfo.width"] = "must be positive"
	}
	if req.ImageInfo.Height <= 0 {
		details["imageInfo.height"] = "must be positive"
	}
	if req.ImageInfo.SizeBytes <= 0 {
		details["imageInfo.size_bytes"] = "must be positive"
	}

	if len(details) > 0 {
		us.logger.Info("Upload request validation failed", "details", details)
		return apperrors.NewBadRequestError("invalid upload image request", details)
	}
	return nil
}

func (us *UploadService) ValidateAndCreatePatient(info *PatientInfo) *models.Patient {
	if info.Age == nil && info.Race == nil && info.Gender == nil &&
		info.History == nil && info.Disease == nil {
		return nil
	}

	patient := &models.Patient{
		Age:     info.Age,
		Race:    info.Race,
		Gender:  info.Gender,
		History: info.History,
		Disease: info.Disease,
	}

	return patient
}

func (us *UploadService) GenerateSignedUploadURL(ctx context.Context, fileName string, contentType string) (string, error) {

	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(30 * time.Minute),
		ContentType: contentType,
		Headers:     []string{fmt.Sprintf("Content-Type:%s", contentType)},
	}

	u, err := us.storageClient.Bucket(us.bucketName).SignedURL(fileName, opts)
	if err != nil {
		return "", apperrors.NewInternalError("failed to generate signed URL", err)
	}
	return u, nil
}

func (us *UploadService) ProcessUpload(ctx context.Context, req *UploadImageRequest) (*SignedUrlResponse, error) {
	// Check creator exists
	CreatorID, ok := ctx.Value("user_id").(string)
	if !ok || CreatorID == "" {
		return nil, apperrors.NewUnauthorizedError("missing user ID in context")
	}

	if err := us.validateUploadRequest(ctx, req); err != nil {
		return nil, err
	}

	var imageID, patientID string
	var fileName string

	err := us.repo.RunTransaction(ctx, func(txCtx context.Context, tx repository.Transaction) error {

		patientID = req.PatientInfo.PatientID

		if patientID == "" {
			patient := us.ValidateAndCreatePatient(&req.PatientInfo)
			if patient != nil {
				patient.CreatedAt = time.Now()
				patient.UpdatedAt = time.Now()

				id, err := tx.Create(string(PatientCollection), patient.ToMap())
				if err != nil {
					return apperrors.NewInternalError("failed to create patient transactionally", err)
				}
				patientID = id
				us.logger.Info("Patient created in transaction", "patientID", patientID)

			}

		} else {
			// Verify patient exists
			_, err := tx.Read(string(PatientCollection), patientID)
			if err != nil {
				return apperrors.NewNotFoundError("patient not found")
			}
		}

		imageID = uuid.New().String()
		fileName = fmt.Sprintf("%s-%s", imageID, req.ImageInfo.FileName)

		image := &models.Image{
			ID:            imageID,
			FileName:      fileName,
			Format:        req.ImageInfo.Format,
			Width:         req.ImageInfo.Width,
			Height:        req.ImageInfo.Height,
			SizeBytes:     req.ImageInfo.SizeBytes,
			CreatorID:     CreatorID,
			PatientID:     patientID,
			WorkspaceID:   req.WorkspaceID,
			OriginPath:    fmt.Sprintf("gs://%s/%s", us.bucketName, fileName),
			ProcessedPath: "",
			Status:        models.StatusUploadWaiting,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		imgID, err := tx.Create(string(ImageCollection), image.ToMap())
		if err != nil {
			return apperrors.NewInternalError("failed to create image record transactionally", err)
		}
		imageID = imgID

		us.logger.Info("Image record created in transaction",
			"image_id", imageID,
			"patient_id", patientID,
			"workspace_id", req.WorkspaceID,
		)
		return nil
	})

	if err != nil {
		us.logger.Error("Transaction failed during upload process", "error", err)
		return nil, err
	}

	//Generate signed URL
	contentType := MimeTypeFromFormat(req.ImageInfo.Format)
	uploadURL, err := us.GenerateSignedUploadURL(ctx, fileName, contentType)
	if err != nil {
		return nil, err
	}

	resp := &SignedUrlResponse{
		UploadURL: uploadURL,
		ImageID:   imageID,
		ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
	}

	return resp, nil
}
