package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/port/cache"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type IndexMap struct {
	Version string                `json:"version"`
	Tiles   map[string]TileOffset `json:"tiles"`
}

type TileOffset struct {
	Offset int64 `json:"offset"`
	Length int64 `json:"length"`
}

type TileServer struct {
	cache       cache.Cache
	keyBuilder  *cache.KeyBuilder
	contentRepo port.ContentRepository
	imageRepo   port.ImageRepository
	storage     port.Storage
}

func NewTileServer(
	cache cache.Cache,
	keyBuilder *cache.KeyBuilder,
	contentRepo port.ContentRepository,
	imageRepo port.ImageRepository,
	storage port.Storage,
) *TileServer {
	return &TileServer{
		cache:       cache,
		keyBuilder:  keyBuilder,
		contentRepo: contentRepo,
		imageRepo:   imageRepo,
		storage:     storage,
	}
}

func (s *TileServer) ServeRequest(ctx context.Context, imageID, objectPath string) (io.ReadCloser, error) {
	requestType := s.determineRequestType(objectPath)

	switch requestType {
	case RequestTypeDZI:
		return s.serveDZI(ctx, imageID)

	case RequestTypeThumbnail:
		return s.serveThumbnail(ctx, imageID)

	case RequestTypeIndexMap:
		return s.serveIndexMap(ctx, imageID)

	case RequestTypeTile:
		return s.serveTile(ctx, imageID, objectPath)

	default:
		return nil, errors.NewBadRequestError(
			fmt.Sprintf("unknown request type for path: %s", objectPath),
			nil,
		)
	}
}

type RequestType int

const (
	RequestTypeUnknown RequestType = iota
	RequestTypeDZI
	RequestTypeThumbnail
	RequestTypeIndexMap
	RequestTypeTile
)

func (s *TileServer) determineRequestType(objectPath string) RequestType {
	lowerPath := strings.ToLower(objectPath)

	switch {
	case strings.HasSuffix(lowerPath, ".dzi"):
		return RequestTypeDZI

	case strings.Contains(lowerPath, "thumbnail") || strings.HasPrefix(lowerPath, "thumb"):
		return RequestTypeThumbnail

	case strings.HasSuffix(lowerPath, "indexmap.json"):
		return RequestTypeIndexMap

	case strings.Contains(objectPath, "/"): // Tile requests: "0/0_0.jpeg", "1/1_2.jpeg"
		return RequestTypeTile

	default:
		return RequestTypeUnknown
	}
}

func (s *TileServer) serveDZI(ctx context.Context, imageID string) (io.ReadCloser, error) {
	image, err := s.getImage(ctx, imageID)
	if err != nil {
		return nil, err
	}

	if image.DziContentID == nil {
		return nil, errors.NewNotFoundError("DZI content not found for image")
	}

	content, err := s.getContentMetadata(ctx, *image.DziContentID)
	if err != nil {
		return nil, err
	}

	return s.storage.Get(ctx, *content)
}

func (s *TileServer) serveThumbnail(ctx context.Context, imageID string) (io.ReadCloser, error) {
	image, err := s.getImage(ctx, imageID)
	if err != nil {
		return nil, err
	}

	if image.ThumbnailContentID == nil {
		return nil, errors.NewNotFoundError("thumbnail content not found for image")
	}

	content, err := s.getContentMetadata(ctx, *image.ThumbnailContentID)
	if err != nil {
		return nil, err
	}

	return s.storage.Get(ctx, *content)
}

func (s *TileServer) serveIndexMap(ctx context.Context, imageID string) (io.ReadCloser, error) {
	image, err := s.getImage(ctx, imageID)
	if err != nil {
		return nil, err
	}

	if image.IndexmapContentID == nil {
		return nil, errors.NewNotFoundError("index map content not found for image")
	}

	content, err := s.getContentMetadata(ctx, *image.IndexmapContentID)
	if err != nil {
		return nil, err
	}

	return s.storage.Get(ctx, *content)
}

func (s *TileServer) serveTile(ctx context.Context, imageID, tilePath string) (io.ReadCloser, error) {
	image, err := s.getImage(ctx, imageID)
	if err != nil {
		return nil, err
	}

	usesArchive := image.ZipTilesContentID != nil
	usesTilesDir := image.TilesContentID != nil

	if usesArchive {
		return s.serveTileFromArchive(ctx, imageID, *image.ZipTilesContentID, tilePath)
	} else if usesTilesDir {
		return s.serveTileFromDirectory(ctx, *image.TilesContentID, tilePath)
	}

	return nil, errors.NewNotFoundError("no tile storage configured for image")
}

func (s *TileServer) serveTileFromDirectory(ctx context.Context, tilesContentID, tilePath string) (io.ReadCloser, error) {
	content, err := s.getContentMetadata(ctx, tilesContentID)
	if err != nil {
		return nil, err
	}

	tileContent := *content
	tileContent.Path = fmt.Sprintf("%s/%s", content.Path, tilePath)

	return s.storage.Get(ctx, tileContent)
}

func (s *TileServer) serveTileFromArchive(ctx context.Context, imageID, archiveContentID, tilePath string) (io.ReadCloser, error) {
	archiveContent, err := s.getContentMetadata(ctx, archiveContentID)
	if err != nil {
		return nil, err
	}

	indexMap, err := s.getIndexMap(ctx, imageID)
	if err != nil {
		return nil, err
	}

	tileKey := strings.TrimSuffix(tilePath, filepath.Ext(tilePath))

	tileOffset, exists := indexMap.Tiles[tileKey]
	if !exists {
		return nil, errors.NewNotFoundError(fmt.Sprintf("tile not found in index map: %s", tilePath))
	}

	return s.storage.GetRange(ctx, *archiveContent, tileOffset.Offset, tileOffset.Length)
}

func (s *TileServer) getImage(ctx context.Context, imageID string) (*model.Image, error) {
	cacheKey := s.keyBuilder.Build("image", "metadata", imageID)

	if val, err := s.cache.Get(ctx, cacheKey); err == nil && val != nil {
		if image, ok := val.(*model.Image); ok {
			return image, nil
		}
	}

	image, err := s.imageRepo.Read(ctx, imageID)
	if err != nil {
		return nil, errors.NewInternalError("failed to read image", err)
	}

	_ = s.cache.Set(ctx, cacheKey, image, 10*time.Minute)

	return image, nil
}

func (s *TileServer) getContentMetadata(ctx context.Context, contentID string) (*model.Content, error) {
	cacheKey := s.keyBuilder.Build("content", "metadata", contentID)

	if val, err := s.cache.Get(ctx, cacheKey); err == nil && val != nil {
		if content, ok := val.(*model.Content); ok {
			return content, nil
		}
	}

	content, err := s.contentRepo.Read(ctx, contentID)
	if err != nil {
		return nil, errors.NewInternalError("failed to read content metadata", err)
	}

	_ = s.cache.Set(ctx, cacheKey, content, 10*time.Minute)

	return content, nil
}

func (s *TileServer) getIndexMap(ctx context.Context, imageID string) (*IndexMap, error) {
	cacheKey := s.keyBuilder.Build("indexmap", imageID)

	if val, err := s.cache.Get(ctx, cacheKey); err == nil && val != nil {
		if indexMap, ok := val.(*IndexMap); ok {
			return indexMap, nil
		}
	}

	image, err := s.getImage(ctx, imageID)
	if err != nil {
		return nil, err
	}

	if image.IndexmapContentID == nil {
		return nil, errors.NewNotFoundError("index map not configured for image")
	}

	indexMapContent, err := s.getContentMetadata(ctx, *image.IndexmapContentID)
	if err != nil {
		return nil, err
	}

	reader, err := s.storage.Get(ctx, *indexMapContent)
	if err != nil {
		return nil, errors.NewInternalError("failed to read index map from storage", err)
	}
	defer reader.Close()

	var indexMap IndexMap
	if err := json.NewDecoder(reader).Decode(&indexMap); err != nil {
		return nil, errors.NewInternalError("failed to parse index map JSON", err)
	}

	_ = s.cache.Set(ctx, cacheKey, &indexMap, 30*time.Minute)

	return &indexMap, nil
}

func (s *TileServer) InvalidateImage(ctx context.Context, imageID string) error {
	patterns := []string{
		s.keyBuilder.BuildPattern("image", "*", imageID),
		s.keyBuilder.BuildPattern("indexmap", imageID),
	}

	for _, pattern := range patterns {
		_, _ = s.cache.DeletePattern(ctx, pattern)
	}

	return nil
}

func (s *TileServer) GetStats(ctx context.Context) (*cache.Stats, error) {
	return s.cache.GetStats(ctx)
}
