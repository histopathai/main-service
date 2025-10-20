package service

import (
	"context"
	"fmt"
	"time"
	apperrors "histopathai/internal/errors"
	"histopathai/internal/repository"
	"histopathai/models"

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
}

type SignedUrlRequest struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	ExpiresIn   int64  `json:"expiresIn"` // seconds, optional
}

type SignedUrlResponse struct {
	UploadURL string `json:"uploadUrl"`
	ExpiresAt int64  `json:"expiresAt"` // unix timestamp
}

type ImageInfo struct {
	FileName  string `json:"filename"`
	Format    string `json:"format"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	SizeBytes int64  `json:"size_bytes"`
}

type PatientInfo struct {
	PatientID string `json:"id,omitempty"` // optional, will be created if empty
	Age       int    `json:"age,omitempty"`
	Gender    string `json:"gender,omitempty"`
	Race      string `json:"race,omitempty"`
	Disease   string `json:"disease,omitempty"`
	History   string `json:"history,omitempty"`
}

type UploadImageRequest struct {
	CreatorID   string      `json:"creator_id"`
	ImageInfo   ImageInfo   `json:"image_info"`
	PatientInfo PatientInfo `json:"patient_info"`
	WorkspaceID string      `json:"workspace_id"`
}

type UploadImageResponse struct {
	ImageID   string `json:"image_id"`
	PatientID string `json:"patient_id"`
	FileURL   string `json:"file_url"`
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
		userRepo:       userRepo,
	}
}

func (us *UploadService) ValidateUploadRequest(ctx context.Context, req *UploadImageRequest) error {
	details := make(map[string]interface{})

	// Check if workspace exists
	exists, err := us.workspaceRepo.Exists(ctx, req.WorkspaceID)
	if err != nil {
		return apperrors.NewInternalError("failed to validate workspace", err, details)
	}
	if !exists {
		return apperrors.NewBadRequestError(fmt.Sprintf("workspace %s does not exist", req.WorkspaceID), details)
	}

	// Check image info
	if req.ImageInfo.FileName == "" {
		details["imageInfo.fileName"] = "required"
	}
	if req.ImageInfo.Format == "" {
		details["imageInfo.format"] = "required"
	} else if !models.SupportedImageFormats[req.ImageInfo.Format] {
		details["imageInfo.format"] = fmt.Sprintf("unsupported format: %s", req.ImageInfo.Format)
	}
	if req.ImageInfo.Width <= 0 {
		details["imageInfo.width"] = "must be positive"
	}
	if req.ImageInfo.Height <= 0 {
		details["imageInfo.height"] = "must be positive"
	}
	if req.ImageInfo.SizeBytes <= 0 {
		details["imageInfo.sizeBytes"] = "must be positive"
	}

	if len(details) > 0 {
		return apperrors.NewBadRequestError("invalid upload image request", details)
	}
	return nil
}

func (us *UploadService) ValidateAndCreatePatient(info *PatientInfo) *models.Patient {
	if info.Age <= 0 && info.Race == "" && info.Gender == "" &&
		info.History == "" && info.Disease == "" {
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

func (us *UploadService) ProcessUpload(ctx context.Context, req *UploadImageRequest) (*UploadImageResponse, error) {
	
	if err := us.ValidateUploadRequest(ctx, req); err != nil {
		return nil, err
	}

	patientId := req.PatientInfo.PatientID

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
		ID:          imageID,
		FileName:    fileName,
		Format:      req.ImageInfo.Format,
		Width:       req.ImageInfo.Width,
		Height:      req.ImageInfo.Height,
		SizeBytes:   req.ImageInfo.SizeBytes,
		PatientID:   patientID,
		WorkspaceID: req.WorkspaceID,
		OriginPath:  fmt.Sprintf("gs://%s/%s", us.bucketName, fileName),
		ProcessedPath: "",
		Status:      models.StatusUploadWaiting,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	imgID, err := us.imgRepo.CreateImage(ctx, image)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to create image record", err)
	}

	fileURL := fmt.Sprintf("gs://%s/%s", us.bucketName, fileName)



