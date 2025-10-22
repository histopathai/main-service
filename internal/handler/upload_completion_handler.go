package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/models"
)

type UploadCompletionHandler struct {
	messageBroker  *repository.MessageBroker
	repo           *repository.MainRepository
	logger         *slog.Logger
	subscriptionID string
	nextTopicID    string
}

func NewUploadCompletionHandler(
	messageBroker *repository.MessageBroker,
	repo *repository.MainRepository,
	logger *slog.Logger,
	subscriptionID string,
	nextTopicID string,
) *UploadCompletionHandler {
	return &UploadCompletionHandler{
		messageBroker:  messageBroker,
		repo:           repo,
		logger:         logger,
		subscriptionID: subscriptionID,
		nextTopicID:    nextTopicID,
	}
}

type ObjectEvent struct {
	Kind           string            `json:"kind"`
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Bucket         string            `json:"bucket"`
	Size           string            `json:"size"`
	ContentType    string            `json:"contentType"`
	Metageneration string            `json:"metageneration"`
	TimeCreated    string            `json:"timeCreated"`
	Updated        string            `json:"updated"`
	MetaData       map[string]string `json:"metadata"`
}

func (h *UploadCompletionHandler) ParsePatientFromMetadata(metadata map[string]string) (*models.Patient, error) {
	patient := &models.Patient{}

	if ageStr, ok := metadata["patient-age"]; ok {
		age, err := strconv.Atoi(ageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid patient-age: %s", ageStr)
		}
		patient.Age = &age
	}
	if gender, ok := metadata["patient-gender"]; ok {
		patient.Gender = &gender
	}
	if race, ok := metadata["patient-race"]; ok {
		patient.Race = &race
	}
	if disease, ok := metadata["patient-disease"]; ok {
		patient.Disease = &disease
	}
	if history, ok := metadata["patient-history"]; ok {
		patient.History = &history
	}
	if patient.Age == nil && patient.Gender == nil &&
		patient.Race == nil && patient.Disease == nil && patient.History == nil {
		return nil, fmt.Errorf("no patient metadata found")
	}

	return patient, nil
}

func (h *UploadCompletionHandler) StartListening(ctx context.Context) error {
	h.logger.Info("Starting upload completion handler", "subscriptionID", h.subscriptionID)
	return h.messageBroker.SubscribeToMessages(ctx, h.subscriptionID, h.handleMessage)
}

func (h *UploadCompletionHandler) handleMessage(ctx context.Context, data []byte) error {
	var event ObjectEvent
	if err := json.Unmarshal(data, &event); err != nil {
		h.logger.Error("Failed to unmarshal message", "error", err)
		return nil // makes ack
	}

	if event.MetaData == nil {
		h.logger.Info("No metadata found in object event, skipping", "object", event.Name)
		return nil // makes ack
	}

	h.logger.Info("Processing upload-status event", "object", event.Name)

	creatorID, ok1 := event.MetaData["creator-id"]
	workspaceID, ok2 := event.MetaData["workspace-id"]

	if ok1 || ok2 {
		h.logger.Info("Required metadata missing, skipping",
			"object", event.Name,
			"creatorID", creatorID,
			"workspaceID", workspaceID,
		)
		return nil // makes ack
	}

	sizeBytes, _ := strconv.ParseInt(event.Size, 10, 64)

	newImage := &models.Image{
		ID:          event.Name,
		FileName:    event.MetaData["image-filename"],
		Format:      event.MetaData["image-format"],
		SizeBytes:   sizeBytes,
		OriginPath:  event.MetaData["image-origin-path"],
		WorkspaceID: workspaceID,
		CreatorID:   creatorID,
		Status:      models.StatusUploaded,
	}

	var err error

	if patientID, ok := event.MetaData["patient-id"]; ok {
		if patientID != "" {
			newImage.PatientID = patientID
			newImage.CreatedAt = time.Now()
			newImage.UpdatedAt = time.Now()
			_, err := h.repo.Create(ctx, string(repository.ImagesCollection), newImage.ToMap())
			if err != nil {
				h.logger.Error("Failed to create image", "error", err)
				return err
			}
		}
	} else {
		var newPatient *models.Patient
		err = h.repo.RunTransaction(ctx, func(txCtx context.Context, tx repository.Transaction) error {

			newPatient, err := h.ParsePatientFromMetadata(event.MetaData)
			if err != nil {
				return err
			}
			newPatient.ID, err = tx.Create(string(repository.PatientsCollection), newPatient.ToMap())
			if err != nil {
				return err
			}
			newImage.PatientID = newPatient.ID
			newImage.CreatedAt = time.Now()
			newImage.UpdatedAt = time.Now()
			_, err = tx.Create(string(repository.ImagesCollection), newImage.ToMap())
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			h.logger.Error("Failed to create patient and image in transaction", "error", err)
			return err
		}
		// Successfully created patient and image
		h.logger.Info("Successfully created patient and image", "patientID", newPatient.ID, "imageID", newImage.ID)
	}
	// Publish message to next topic for further processing

	messageData, err := json.Marshal(newImage)
	if err != nil {
		h.logger.Error("Failed to marshal image for publishing", "error", err)
		return err
	}
	_, err = h.messageBroker.PublishMessage(ctx, h.nextTopicID, messageData)
	if err != nil {
		h.logger.Error("Failed to publish message to next topic", "error", err)
		return err
	}
	h.logger.Info("Published message to next topic", "topicID", h.nextTopicID, "imageID", newImage.ID)

	return nil
}
