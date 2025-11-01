package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/api/http/validator"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ImageHandler struct {
	imageService *service.ImageService
	validator    *validator.RequestValidator
	BaseHandler  // Embed the BaseHandler
}

func NewImageHandler(imageService *service.ImageService, validator *validator.RequestValidator, logger *slog.Logger) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
		validator:    validator,
		BaseHandler:  BaseHandler{logger: logger},
	}
}

// UploadImage godoc
// @Summary Upload a new image
// @Description Upload a new image with the provided details
// @Tags Images
// @Accept json
// @Produce json
// @Param        request body request.UploadImageRequest true "Image upload request"
// @Success 201 {object} response.UploadImageResponse "Image uploaded successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /images [post]
func (ih *ImageHandler) UploadImage(c *gin.Context) {
	creator_id, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ih.handleError(c, err)
		return
	}

	var req request.UploadImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ih.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	if err := ih.validator.ValidateStruct(&req); err != nil {
		ih.handleError(c, err)
		return
	}

	// DTO -> Service Input
	input := service.UploadImageInput{
		PatientID:   req.PatientID,
		CreatorID:   creator_id,
		ContentType: req.ContentType,
		Name:        req.Name,
		Format:      req.Format,
		Width:       req.Width,
		Height:      req.Height,
		Size:        req.Size,
	}

	signed_url, err := ih.imageService.UploadImage(c.Request.Context(), &input)
	if err != nil {
		ih.handleError(c, err)
		return
	}

	respPayload := response.UploadImagePayload{
		UploadURL: *signed_url,
		Message:   "Use this URL to upload the image via a PUT request.",
	}
	ih.response.Created(c, respPayload)
}

// GetImageByID godoc
// @Summary Get image by ID
// @Description Retrieve image details by its ID
// @Tags Images
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Success 200 {object} response.ImageDataResponse "Image retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "Image not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /images/{image_id} [get]
func (ih *ImageHandler) GetImageByID(c *gin.Context) {
	image_id := c.Param("image_id")
	if image_id == "" {
		ih.handleError(c, errors.NewValidationError("image_id is required", nil))
		return
	}

	image, err := ih.imageService.GetImageByID(c.Request.Context(), image_id)
	if err != nil {
		ih.handleError(c, err)
		return
	}

	imageResp := response.NewImageResponse(image)
	ih.response.Success(c, http.StatusOK, imageResp)
}

// ListImageByPatientID godoc
// @Summary List images by Patient ID
// @Description Retrieve a list of images associated with a specific Patient ID
// @Tags Images
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.ImageListResponse "Images retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized
// @Security BearerAuth
// @Router /patients/{patient_id}/images [get]
func (ih *ImageHandler) ListImageByPatientID(c *gin.Context) {
	patient_id := c.Param("patient_id")
	if patient_id == "" {
		ih.handleError(c, errors.NewValidationError("patient_id is required", nil))
		return
	}

	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		ih.handleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	pagination := query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := ih.imageService.ListImageByPatientID(c.Request.Context(), patient_id, &pagination)
	if err != nil {
		ih.handleError(c, err)
		return
	}

	imageResponses := make([]response.ImageResponse, len(result.Data))
	for i, img := range result.Data {
		imageResponses[i] = *response.NewImageResponse(img)
	}

	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
		Total:   result.Total,
	}

	ih.response.SuccessList(c, imageResponses, paginationResp)
}

// DeleteImage godoc
// @Summary Delete image by ID
// @Description Delete an image using its ID
// @Tags Images
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Success 204 {string} string "Image deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "Image not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /images/{image_id} [delete]
func (ih *ImageHandler) DeleteImage(c *gin.Context) {
	image_id := c.Param("image_id")
	if image_id == "" {
		ih.handleError(c, errors.NewValidationError("image_id is required", nil))
		return
	}

	err := ih.imageService.DeleteImageByID(c.Request.Context(), image_id)
	if err != nil {
		ih.handleError(c, err)
		return
	}

	ih.logger.Info("Image deleted", slog.String("image_id", image_id))
	ih.response.NoContent(c)
}
