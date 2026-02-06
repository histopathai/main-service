package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/handler/helper"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	validator "github.com/histopathai/main-service/internal/shared/query"
)

type ImageHandler struct {
	helper.BaseHandler
	IQuery     port.ImageQuery
	IUseCase   port.ImageUseCase
	IValidator *validator.Validator
}

func NewImageHandler(query port.ImageQuery, useCase port.ImageUseCase, logger *slog.Logger) *ImageHandler {
	return &ImageHandler{
		IQuery:      query,
		IUseCase:    useCase,
		IValidator:  validator.NewValidator(fields.NewImageFieldSet()),
		BaseHandler: helper.NewBaseHandler(logger),
	}
}

// UploadImage godoc
// @Summary Upload a new image
// @Tags Images
// @Accept json
// @Produce json
// @Param request body request.UploadImageRequest true "Image upload request"
// @Success 201 {object} response.UploadImageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images [post]
func (ih *ImageHandler) UploadImage(c *gin.Context) {
	creatorID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	var req request.UploadImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ih.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// DTO -> Command
	contents := make([]struct {
		ContentType string
		Name        string
		Size        int64
	}, len(req.Contents))
	for i, content := range req.Contents {
		contents[i] = struct {
			ContentType string
			Name        string
			Size        int64
		}{
			ContentType: content.ContentType,
			Name:        content.Name,
			Size:        content.Size,
		}
	}

	cmd := command.UploadImageCommand{
		CreateEntityCommand: command.CreateEntityCommand{
			Name:       req.Name,
			EntityType: vobj.EntityTypeImage.String(),
			CreatorID:  creatorID,
			ParentID:   req.Parent.ID,
			ParentType: req.Parent.Type,
		},
		Format:   req.Format,
		Width:    req.Width,
		Height:   req.Height,
		Contents: contents,
		WsID:     req.WsID,
	}

	errDetails, ok := cmd.Validate()
	if !ok {
		ih.HandleError(c, errors.NewValidationError("invalid command payload", errDetails))
		return
	}

	payload, err := ih.IUseCase.Upload(c.Request.Context(), cmd)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	// Command -> Response
	respPayload := make([]response.UploadImagePayload, len(payload))
	for i, p := range payload {
		respPayload[i] = response.UploadImagePayload{
			UploadURL: p.URL,
			Headers:   p.Headers,
			Message:   "Use this URL and Headers to upload the image via a PUT request.",
		}
	}

	ih.Response.Success(c, http.StatusCreated, respPayload)
}

// Get godoc
// @Summary Get image by ID
// @Description Retrieve image details by its ID
// @Tags Images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Success 200 {object} response.ImageDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/{id} [get]
func (ih *ImageHandler) Get(c *gin.Context) {
	image, err := ih.IQuery.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	imageResp := response.NewImageResponse(image)
	ih.Response.Success(c, http.StatusOK, imageResp)
}

func (ih *ImageHandler) List(c *gin.Context) {
	var req request.ListRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		ih.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ih.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	ih.ApplyVisibilityFilters(c, &spec)

	if err := ih.IValidator.ValidateSpec(spec); err != nil {
		ih.HandleError(c, err)
		return
	}

	images, err := ih.IQuery.List(c.Request.Context(), spec)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.Success(c, http.StatusOK, response.NewImageListResponse(images))
}

// GetByParentID godoc
// @Summary Get images by Parent ID
// @Description Get images belonging to a specific parent with optional filtering, sorting, and pagination
// @Tags Images
// @Accept json
// @Produce json
// @Param parent_id path string true "Parent ID"
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name, format)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.ImageListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/parent/{parent_id} [get]
func (ih *ImageHandler) GetByParentID(c *gin.Context) {
	var req request.ListRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		ih.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ih.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	ih.ApplyVisibilityFilters(c, &spec)

	if err := ih.IValidator.ValidateSpec(spec); err != nil {
		ih.HandleError(c, err)
		return
	}

	parentID := c.Param("parent_id")
	if parentID == "" {
		parentID = c.Param("id")
	}

	images, err := ih.IQuery.GetByParentID(c.Request.Context(), spec, parentID)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.Success(c, http.StatusOK, response.NewImageListResponse(images))
}

// GetByWorkspaceID godoc
// @Summary Get images by Workspace ID
// @Description Get images belonging to a specific workspace with optional filtering, sorting, and pagination
// @Tags Images
// @Accept json
// @Produce json
// @Param workspace_id path string true "Workspace ID"
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name, format)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.ImageListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/workspace/{workspace_id} [get]
func (ih *ImageHandler) GetByWorkspaceID(c *gin.Context) {
	var req request.ListRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		ih.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ih.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	ih.ApplyVisibilityFilters(c, &spec)

	if err := ih.IValidator.ValidateSpec(spec); err != nil {
		ih.HandleError(c, err)
		return
	}

	images, err := ih.IQuery.GetByWsID(c.Request.Context(), spec, c.Param("workspace_id"))
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.Success(c, http.StatusOK, response.NewImageListResponse(images))
}

// Count godoc
// @Summary Count images
// @Description Count images with optional filters via query parameters
// @Tags Images
// @Accept json
// @Produce json
// @Success 200 {object} response.CountResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/count [get]
func (ih *ImageHandler) Count(c *gin.Context) {
	var req request.ListRequest

	// Bind query parameters (optional for count)
	if err := c.ShouldBindQuery(&req); err != nil {
		ih.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	if err := ih.IValidator.ValidateSpec(spec); err != nil {
		ih.HandleError(c, err)
		return
	}

	count, err := ih.IQuery.Count(c.Request.Context(), spec)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	respPayload := response.CountResponse{Count: count}
	ih.Response.Success(c, http.StatusOK, respPayload)
}

// Update godoc
// @Summary Update images
// @Tags Images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Param request body request.UpdateImageRequest true "Image update request"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/{id} [put]
func (ih *ImageHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request.UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ih.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	var magnification *vobj.OpticalMagnification
	if req.Magnification != nil {
		magnification = &vobj.OpticalMagnification{
			Objective:         req.Magnification.Objective,
			NativeLevel:       req.Magnification.NativeLevel,
			ScanMagnification: req.Magnification.ScanMagnification,
		}
	}

	cmd := command.UpdateImageCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{
			ID:        id,
			Name:      req.Name,
			CreatorID: req.CreatorID,
		},
		Width:         req.Width,
		Height:        req.Height,
		Magnification: magnification,
	}

	errDetails, ok := cmd.Validate()
	if !ok {
		ih.HandleError(c, errors.NewValidationError("invalid command payload", errDetails))
		return
	}

	err := ih.IUseCase.Update(c.Request.Context(), cmd)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.NoContent(c)
}

// Transfer godoc
// @Summary Transfer image to a new patient
// @Tags Images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Param patient_id path string true "New Patient ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/{id}/transfer/{patient_id} [put]
func (ih *ImageHandler) Transfer(c *gin.Context) {
	imageID := c.Param("id")
	patientID := c.Param("patient_id")

	cmd := command.TransferCommand{
		ID:         imageID,
		NewParent:  patientID,
		ParentType: vobj.EntityTypePatient.String(),
	}

	err := ih.IUseCase.Transfer(c.Request.Context(), cmd)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.NoContent(c)
}

// TransferMany godoc
// @Summary Batch transfer images to a new patient
// @Tags Images
// @Accept json
// @Produce json
// @Param ids query []string true "Image IDs"
// @Param patient_id path string true "Target Patient ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/transfer-many/{patient_id} [put]
func (ih *ImageHandler) TransferMany(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		ih.HandleError(c, errors.NewValidationError("ids parameter is required", nil))
		return
	}

	patientID := c.Param("patient_id")

	cmd := command.TransferManyCommand{
		IDs:        ids,
		NewParent:  patientID,
		ParentType: vobj.EntityTypePatient.String(),
	}

	err := ih.IUseCase.TransferMany(c.Request.Context(), cmd)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.NoContent(c)
}

// SoftDelete godoc
// @Summary Soft delete image by ID
// @Tags Images
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/{id}/soft-delete [delete]
func (ih *ImageHandler) SoftDelete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ih.HandleError(c, errors.NewValidationError("id parameter is required", nil))
		return
	}

	err := ih.IQuery.SoftDelete(c.Request.Context(), id)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.NoContent(c)
}

// SoftDeleteMany godoc
// @Summary Batch soft delete images by IDs
// @Tags Images
// @Accept json
// @Produce json
// @Param ids query []string true "Image IDs"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images/soft-delete-many [delete]
func (ih *ImageHandler) SoftDeleteMany(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		ih.HandleError(c, errors.NewValidationError("ids parameter is required", nil))
		return
	}

	err := ih.IQuery.SoftDeleteMany(c.Request.Context(), ids)
	if err != nil {
		ih.HandleError(c, err)
		return
	}

	ih.Response.NoContent(c)
}
