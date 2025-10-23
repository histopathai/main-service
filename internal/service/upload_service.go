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
	SizeBytes int64  `json:"size_bytes"`
}

type NewPatientInfo struct {
	Age     *int    `json:"age,omitempty"`
	Gender  *string `json:"gender,omitempty"`
	Race    *string `json:"race,omitempty"`
	Disease *string `json:"disease,omitempty"`
	History *string `json:"history,omitempty"`
}

type UploadImageRequest struct {
	ImageInfo   ImageInfo      `json:"image_info"`
	PatientInfo NewPatientInfo `json:"new_patient_info"`
	PatientID   *string        `json:"patient_id,omitempty"`
	WorkspaceID string         `json:"workspace_id"`
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

func (us *UploadService) validateUploadRequest(ctx context.Context, req *UploadImageRequest) (*models.Image, *models.Patient, error) {
	// Scope variables
	var patient *models.Patient
	var image *models.Image
	var err error
	details := make(map[string]interface{})

	// Check workspace exists
	wr, err := us.repo.Read(ctx, string(WorkspaceCollection), req.WorkspaceID)
	if err != nil {
		return nil, nil, apperrors.NewInternalError("failed to read workspace during upload validation", err)
	}
	if wr == nil {
		details["workspace"] = "required"
	} else {
		w := &models.Workspace{}
		w.FromMap(wr)
		if w.AnnotationTypeID == "" {
			details["workspace"] = "workspace must have an annotation type assigned. Assign an annotation type before uploading images."
		}
	}

	// Patient id or New patient info must be provided
	if req.PatientID != nil {
		// Verify patient exists
		p, err := us.repo.Read(ctx, string(PatientCollection), *req.PatientID)
		if err != nil {
			return nil, nil, apperrors.NewInternalError("failed to read patient during upload validation", err)
		}
		if p == nil {
			details["patient_id"] = "patient not found"
		} else {
			patient = &models.Patient{}
			patient.FromMap(p)
		}

	} else {
		// Verify at least one patient detail is provided
		if req.PatientInfo.Age == nil && req.PatientInfo.Race == nil && req.PatientInfo.Gender == nil &&
			req.PatientInfo.History == nil && req.PatientInfo.Disease == nil {
			details["patient_info"] = "either patient ID or at least one patient detail must be provided"
			details["patient_info.age"] = "optional"
			details["patient_info.gender"] = "optional"
			details["patient_info.race"] = "optional"
			details["patient_info.disease"] = "optional"
			details["patient_info.history"] = "optional"
			details["patient_info.patient_id"] = "optional when patient details are provided"
		} else {
			// Create new patient object
			patient = &models.Patient{
				ID:      "",
				Age:     req.PatientInfo.Age,
				Gender:  req.PatientInfo.Gender,
				Race:    req.PatientInfo.Race,
				Disease: req.PatientInfo.Disease,
				History: req.PatientInfo.History,
			}
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

	if req.ImageInfo.SizeBytes <= 0 {
		details["imageInfo.size_bytes"] = "must be positive"
	}

	if len(details) > 0 {
		return nil, nil, apperrors.NewValidationError("invalid upload image request", details)
	}

	// Create image object
	image = &models.Image{}

	uuid := uuid.New().String()

	image.FileName = req.ImageInfo.FileName
	image.Format = req.ImageInfo.Format
	image.SizeBytes = req.ImageInfo.SizeBytes
	image.WorkspaceID = req.WorkspaceID
	image.ID = fmt.Sprintf("%s-%s", uuid, req.ImageInfo.FileName)
	image.OriginPath = fmt.Sprintf("gs://%s/%s", us.bucketName, image.ID)
	image.PatientID = patient.ID

	return image, patient, nil
}

func (us *UploadService) GenerateSignedUploadURL(ctx context.Context, contentType string, image *models.Image, patient *models.Patient) (string, error) {

	headers := []string{
		"Content-Type:" + contentType,
		"x-goog-meta-image-id:" + image.ID,
		"x-goog-meta-workspace-id:" + image.WorkspaceID,
		"x-goog-meta-creator-id:" + image.CreatorID,
		"x-goog-meta-image-format:" + image.Format,
		"x-goog-meta-image-size-bytes:" + fmt.Sprintf("%d", image.SizeBytes),
		"x-goog-meta-image-filename:" + image.FileName,
		"x-goog-meta-image-origin-path:" + image.OriginPath,
	}

	if image.PatientID != "" {
		headers = append(headers, "x-goog-meta-patient-id:"+image.PatientID)
	} else {
		if patient.Age != nil {
			headers = append(headers, "x-goog-meta-patient-age:"+fmt.Sprintf("%d", *patient.Age))
		}
		if patient.Gender != nil {
			headers = append(headers, "x-goog-meta-patient-gender:"+*patient.Gender)
		}
		if patient.Race != nil {
			headers = append(headers, "x-goog-meta-patient-race:"+*patient.Race)
		}
		if patient.Disease != nil {
			headers = append(headers, "x-goog-meta-patient-disease:"+*patient.Disease)
		}
		if patient.History != nil {
			headers = append(headers, "x-goog-meta-patient-history:"+*patient.History)
		}
	}

	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(30 * time.Minute),
		ContentType: contentType,
		Headers:     headers,
	}

	u, err := us.storageClient.Bucket(us.bucketName).SignedURL(image.ID, opts)
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
	// Validate request
	image, patient, err := us.validateUploadRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	image.CreatorID = CreatorID

	contentType := MimeTypeFromFormat(req.ImageInfo.Format)
	UploadURL, err := us.GenerateSignedUploadURL(ctx, contentType, image, patient)
	if err != nil {
		return nil, err
	}

	return &SignedUrlResponse{
		ImageID:   image.ID,
		UploadURL: UploadURL,
		ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
	}, nil
}
