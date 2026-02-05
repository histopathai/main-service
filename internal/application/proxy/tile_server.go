package proxy

import (
	"context"
	"encoding/binary"
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
	Version interface{}           `json:"version"`
	Entries []IndexMapEntry       `json:"entries"`
	Tiles   map[string]TileOffset `json:"-"`
}

type IndexMapEntry struct {
	Name             string `json:"name"`
	Offset           int64  `json:"offset"`
	CompressedSize   int64  `json:"compressed_size"`
	UncompressedSize int64  `json:"uncompressed_size"`
	Method           int    `json:"method"`
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

	// The offset in the index map typically points to the ZIP Local File Header.
	// We need to read this header to find the start of the actual data.
	// Local File Header fixed size is 30 bytes.
	headerReader, err := s.storage.GetRange(ctx, *archiveContent, tileOffset.Offset, 30)
	if err != nil {
		return nil, errors.NewInternalError("failed to verify zip header", err)
	}

	header, err := io.ReadAll(headerReader)
	_ = headerReader.Close()
	if err != nil {
		return nil, errors.NewInternalError("failed to read zip header bytes", err)
	}

	dataOffset := tileOffset.Offset

	// Check signature: 0x04034b50 (PK\x03\x04)
	if len(header) == 30 && header[0] == 0x50 && header[1] == 0x4b && header[2] == 0x03 && header[3] == 0x04 {
		// Offset 26: Filename length (2 bytes)
		// Offset 28: Extra field length (2 bytes)
		nameLen := int64(binary.LittleEndian.Uint16(header[26:28]))
		extraLen := int64(binary.LittleEndian.Uint16(header[28:30]))

		// Total header size = 30 + nameLen + extraLen
		dataOffset += 30 + nameLen + extraLen
	}

	return s.storage.GetRange(ctx, *archiveContent, dataOffset, tileOffset.Length)
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

	// Populate lookup map
	indexMap.Tiles = make(map[string]TileOffset)
	for _, entry := range indexMap.Entries {
		// Normalize key: remove "image/" prefix if exists, and remove extension
		// Entry name example: "image/image_files/12/2_0.jpg"
		// Request path: "image_files/12/2_0.jpg" -> Key: "image_files/12/2_0"

		key := entry.Name
		if strings.HasPrefix(key, "image/") {
			key = strings.TrimPrefix(key, "image/")
		}

		// Remove extension
		ext := filepath.Ext(key)
		if ext != "" {
			key = strings.TrimSuffix(key, ext)
		}

		indexMap.Tiles[key] = TileOffset{
			Offset: entry.Offset,
			Length: entry.CompressedSize, // Use CompressedSize for reading from zip
		}
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
