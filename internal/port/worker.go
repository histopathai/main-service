package port

import (
	"context"

	model "github.com/histopathai/main-service/internal/domain/model"
	vobj "github.com/histopathai/main-service/internal/domain/vobj"
)

type ImageProcessingWorker interface {
	ProcessImage(ctx context.Context, content model.Content, processingVersion vobj.ProcessingVersion) error
}
