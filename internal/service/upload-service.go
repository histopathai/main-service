package service

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/histopathai/models"
)

type UploadService struct {
	storageClient *storage.Client
	bucketName    string
	imgRepo       *repository.ImageRepository
	patientRepo   *repository.PatientRepository
	workspaceRepo *repository.WorkspaceRepository
	userRepo      *repository.UserRepository
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
	CreatorID   string      `json:"creator_id"`
	ImageInfo   ImageInfo   `json:"image_info"`
	PatientInfo PatientInfo `json:"patient_info"`
	WorkspaceID string      `json:"workspace_id"`
}

func MimeeTypeFromFormat(format string) string {
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
	imgRepo *repository.ImageRepository,
	patientRepo *repository.PatientRepository,
	workspaceRepo *repository.WorkspaceRepository,
	userRepo *repository.UserRepository,
) *UploadService {
	return &UploadService{
		storageClient: storageClient,
		bucketName:    bucketName,
		imgRepo:       imgRepo,
		patientRepo:   patientRepo,
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (us *UploadService) ValidateUploadRequest(ctx context.Context, req *UploadImageRequest) error {
	details := make(map[string]interface{})
	// Check creator exists
	exists, err := us.userRepo.Exists(ctx, req.CreatorID)
	if err != nil {
		return apperrors.NewInternalError("failed to validate creator", err)
	}
	if !exists {
		return apperrors.NewBadRequestError(fmt.Sprintf("creator %s does not exist", req.CreatorID), details)
	}

	// Check if workspace exists
	exists, err = us.workspaceRepo.Exists(ctx, req.WorkspaceID)
	if err != nil {
		return apperrors.NewInternalError("failed to validate workspace", err)
	}
	if !exists {
		return apperrors.NewBadRequestError(fmt.Sprintf("workspace %s does not exist", req.WorkspaceID), details)
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

	if err := us.ValidateUploadRequest(ctx, req); err != nil {
		return nil, err
	}

	patientID := req.PatientInfo.PatientID

	if patientID == "" {
		patient := us.ValidateAndCreatePatient(&req.PatientInfo)
		if patient != nil {
			id, err := us.patientRepo.CreatePatient(ctx, patient)
			if err != nil {
				return nil, apperrors.NewInternalError("failed to create patient", err)
			}
			patientID = id
		}
	} else {
		// Patient ID varsa, varlığını kontrol et
		exists, err := us.patientRepo.Exists(ctx, patientID)
		if err != nil {
			return nil, apperrors.NewInternalError("failed to check patient existence", err)
		}
		if !exists {
			return nil, apperrors.NewNotFoundError("patient not found")
		}
	}

	imageID := uuid.New().String()
	fileName := fmt.Sprintf("%s-%s", imageID, req.ImageInfo.FileName)

	image := &models.Image{
		ID:            imageID,
		FileName:      fileName,
		Format:        req.ImageInfo.Format,
		Width:         req.ImageInfo.Width,
		Height:        req.ImageInfo.Height,
		SizeBytes:     req.ImageInfo.SizeBytes,
		PatientID:     patientID,
		WorkspaceID:   req.WorkspaceID,
		OriginPath:    fmt.Sprintf("gs://%s/%s", us.bucketName, fileName),
		ProcessedPath: "",
		Status:        models.StatusUploadWaiting,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	imgID, err := us.imgRepo.CreateImage(ctx, image)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to create image record", err)
	}

	contentType := MimeeTypeFromFormat(req.ImageInfo.Format)
	uploadURL, err := us.GenerateSignedUploadURL(ctx, fileName, contentType)
	if err != nil {
		return nil, err
	}

	resp := &SignedUrlResponse{
		UploadURL: uploadURL,
		ImageID:   imgID,
		ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
	}

	return resp, nil
}
