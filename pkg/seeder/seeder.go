// pkg/seeder/seeder.go
package seeder

import (
	"context"
	"log/slog"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
)

type Seeder struct {
	repos  *repository.Repositories
	logger *slog.Logger
}

func NewSeeder(repos *repository.Repositories, logger *slog.Logger) *Seeder {
	return &Seeder{
		repos:  repos,
		logger: logger,
	}
}

func (s *Seeder) Seed(ctx context.Context) error {
	s.logger.Info("Starting database seeding...")

	// Seed annotation types
	if err := s.seedAnnotationTypes(ctx); err != nil {
		return err
	}

	// Seed workspaces
	if err := s.seedWorkspaces(ctx); err != nil {
		return err
	}

	s.logger.Info("Database seeding completed successfully")
	return nil
}

func (s *Seeder) seedAnnotationTypes(ctx context.Context) error {
	s.logger.Info("Seeding annotation types...")

	annotationTypes := []struct {
		name                  string
		description           string
		scoreEnabled          bool
		scoreName             *string
		scoreRange            *[2]float64
		classificationEnabled bool
		classList             []string
	}{
		{
			name:                  "Tumor Classification",
			description:           "Classification of tumor regions",
			scoreEnabled:          false,
			classificationEnabled: true,
			classList:             []string{"Benign", "Malignant"},
		},
		{
			name:                  "Tumor Grade",
			description:           "Grading of tumor severity",
			scoreEnabled:          true,
			scoreName:             stringPtr("Grade"),
			scoreRange:            &[2]float64{1.0, 5.0},
			classificationEnabled: false,
		},
		{
			name:                  "Necrosis",
			description:           "Identification of necrotic tissue",
			scoreEnabled:          false,
			classificationEnabled: false,
		},
		{
			name:                  "Inflammation Level",
			description:           "Assessment of inflammation severity",
			scoreEnabled:          true,
			scoreName:             stringPtr("Severity"),
			scoreRange:            &[2]float64{0.0, 10.0},
			classificationEnabled: false,
		},
	}

	creatorID := "system"

	for _, at := range annotationTypes {
		// Check if already exists
		existing, err := s.repos.AnnotationTypeRepo.FindByName(ctx, at.name)
		if err == nil && existing != nil {
			s.logger.Info("Annotation type already exists, skipping",
				slog.String("name", at.name))
			continue
		}

		desc := at.description
		annotationType := &model.AnnotationType{
			CreatorID:             creatorID,
			Name:                  at.name,
			Description:           &desc,
			ScoreEnabled:          at.scoreEnabled,
			ScoreName:             at.scoreName,
			ScoreRange:            at.scoreRange,
			ClassificationEnabled: at.classificationEnabled,
			ClassList:             at.classList,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		created, err := s.repos.AnnotationTypeRepo.Create(ctx, annotationType)
		if err != nil {
			s.logger.Error("Failed to create annotation type",
				slog.String("name", at.name),
				slog.String("error", err.Error()))
			return err
		}

		s.logger.Info("Created annotation type",
			slog.String("id", created.ID),
			slog.String("name", created.Name))
	}

	return nil
}

func (s *Seeder) seedWorkspaces(ctx context.Context) error {
	s.logger.Info("Seeding workspaces...")

	workspaces := []struct {
		name         string
		organType    string
		organization string
		description  string
		license      string
		resourceURL  *string
		releaseYear  *int
	}{
		{
			name:         "Brain Tumor Research",
			organType:    "brain",
			organization: "Neuro Research Institute",
			description:  "A comprehensive workspace for brain tumor histopathology analysis",
			license:      "CC BY 4.0",
			releaseYear:  intPtr(2024),
		},
		{
			name:         "Lung Cancer Study",
			organType:    "lung",
			organization: "Pulmonary Oncology Center",
			description:  "Workspace dedicated to lung cancer tissue analysis and classification",
			license:      "CC BY-NC 4.0",
			releaseYear:  intPtr(2024),
		},
		{
			name:         "Liver Pathology Database",
			organType:    "liver",
			organization: "Hepatology Research Lab",
			description:  "Collection of liver tissue samples for various pathological conditions",
			license:      "CC BY-SA 4.0",
			releaseYear:  intPtr(2023),
		},
	}

	creatorID := "system"

	for _, ws := range workspaces {
		// Check if already exists
		existing, err := s.repos.WorkspaceRepo.FindByName(ctx, ws.name)
		if err == nil && existing != nil {
			s.logger.Info("Workspace already exists, skipping",
				slog.String("name", ws.name))
			continue
		}

		workspace := &model.Workspace{
			CreatorID:    creatorID,
			Name:         ws.name,
			OrganType:    ws.organType,
			Organization: ws.organization,
			Description:  ws.description,
			License:      ws.license,
			ResourceURL:  ws.resourceURL,
			ReleaseYear:  ws.releaseYear,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		created, err := s.repos.WorkspaceRepo.Create(ctx, workspace)
		if err != nil {
			s.logger.Error("Failed to create workspace",
				slog.String("name", ws.name),
				slog.String("error", err.Error()))
			return err
		}

		s.logger.Info("Created workspace",
			slog.String("id", created.ID),
			slog.String("name", created.Name))
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
