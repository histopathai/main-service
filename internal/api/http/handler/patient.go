package handler

import (
	"io"
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

type PatientHandler struct {
	helper.BaseHandler

	PQuery     port.PatientQuery
	PUseCase   port.PatientUseCase
	PValidator *validator.Validator
}

func NewPatientHandler(patientQuery port.PatientQuery, useCase port.PatientUseCase, logger *slog.Logger) *PatientHandler {
	return &PatientHandler{
		PQuery:      patientQuery,
		PUseCase:    useCase,
		PValidator:  validator.NewValidator(fields.NewPatientFieldSet()),
		BaseHandler: helper.NewBaseHandler(logger),
	}
}

// Create godoc
// @Summary Create a new patient
// @Tags Patients
// @Accept json
// @Produce json
// @Param request body request.CreatePatientRequest true "Patient creation request"
// @Success 201 {object} response.PatientDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients [post]
func (ph *PatientHandler) Create(c *gin.Context) {
	creatorID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	var req request.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.HandleError(c, errors.NewValidationError("invalid request payload",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	// DTO -> Command
	cmd := command.CreatePatientCommand{
		CreateEntityCommand: command.CreateEntityCommand{
			Name:       req.Name,
			EntityType: vobj.EntityTypePatient.String(),
			CreatorID:  creatorID,
			ParentID:   req.Parent.ID,
			ParentType: vobj.ParentTypeWorkspace.String(),
		},
		Age:     req.Age,
		Race:    req.Race,
		Gender:  req.Gender,
		Disease: req.Disease,
		Subtype: req.Subtype,
		Grade:   req.Grade,
		History: req.History,
	}

	errDetails, ok := cmd.Validate()
	if !ok {
		ph.HandleError(c, errors.NewValidationError("Invalid request", errDetails))
		return
	}

	patient, err := ph.PUseCase.Create(c.Request.Context(), cmd)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.Created(c, response.NewPatientResponse(patient))
}

// Get godoc
// @Summary Get patient by ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param id path string true "Patient ID"
// @Success 200 {object} response.PatientDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/{id} [get]
func (ph *PatientHandler) Get(c *gin.Context) {
	patient, err := ph.PQuery.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.Success(c, http.StatusOK, response.NewPatientResponse(patient))
}

// List godoc
// @Summary List patients
// @Description List patients with optional filtering, sorting, and pagination via query parameters
// @Tags Patients
// @Accept json
// @Produce json
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name, age, disease)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.PatientListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients [get]
func (ph *PatientHandler) List(c *gin.Context) {
	var req request.ListRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		ph.HandleError(c, errors.NewValidationError("invalid query parameters",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ph.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	ph.ApplyVisibilityFilters(c, &spec)

	if err := ph.PValidator.ValidateSpec(spec); err != nil {
		ph.HandleError(c, err)
		return
	}

	result, err := ph.PQuery.List(c.Request.Context(), spec)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	patientResponses := make([]response.PatientResponse, len(result.Data))
	for i, patient := range result.Data {
		patientResponses[i] = *response.NewPatientResponse(patient)
	}

	ph.Response.SuccessList(c, patientResponses, paginationResp)
}

// GetByParentID godoc
// @Summary Get patients by workspace ID
// @Description Get patients belonging to a specific workspace with optional filtering, sorting, and pagination
// @Tags Patients
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name, age, disease)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.PatientListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces/{parent_id}/patients [get]
func (ph *PatientHandler) GetByParentID(c *gin.Context) {
	var req request.ListRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		ph.HandleError(c, errors.NewValidationError("invalid query parameters",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ph.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	ph.ApplyVisibilityFilters(c, &spec)

	if err := ph.PValidator.ValidateSpec(spec); err != nil {
		ph.HandleError(c, err)
		return
	}

	result, err := ph.PQuery.GetByParentID(c.Request.Context(), spec, c.Param("id"))
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	patientResponses := make([]response.PatientResponse, len(result.Data))
	for i, patient := range result.Data {
		patientResponses[i] = *response.NewPatientResponse(patient)
	}

	ph.Response.SuccessList(c, patientResponses, paginationResp)
}

// Count godoc
// @Summary Count patients
// @Description Count patients with optional filters via query parameters
// @Tags Patients
// @Accept json
// @Produce json
// @Success 200 {object} response.CountResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/count [get]
func (ph *PatientHandler) Count(c *gin.Context) {
	var req request.ListRequest
	var spec validator.Specification

	// Try to bind query parameters (optional for count)
	if err := c.ShouldBindQuery(&req); err != nil && err != io.EOF {
		ph.HandleError(c, errors.NewValidationError("invalid query parameters",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ph.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := ph.PValidator.ValidateSpec(spec); err != nil {
		ph.HandleError(c, err)
		return
	}

	count, err := ph.PQuery.Count(c.Request.Context(), spec)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.Success(c, http.StatusOK, response.CountResponse{Count: count})
}

// Update godoc
// @Summary Update patient by ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param id path string true "Patient ID"
// @Param request body request.UpdatePatientRequest true "Patient update request"
// @Success 204 "Patient updated successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/{id} [put]
func (ph *PatientHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.HandleError(c, errors.NewValidationError("invalid request payload",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	// DTO -> Command
	cmd := command.UpdatePatientCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{
			ID:        id,
			Name:      req.Name,
			CreatorID: req.CreatorID,
		},
		Age:     req.Age,
		Gender:  req.Gender,
		Race:    req.Race,
		Disease: req.Disease,
		Subtype: req.Subtype,
		Grade:   req.Grade,
		History: req.History,
	}

	err := ph.PUseCase.Update(c.Request.Context(), cmd)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.NoContent(c)
}

// Transfer godoc
// @Summary Transfer patient to another workspace
// @Tags Patients
// @Accept json
// @Produce json
// @Param id path string true "Patient ID"
// @Param workspace_id path string true "Target Workspace ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/{id}/transfer/{workspace_id} [put]
func (ph *PatientHandler) Transfer(c *gin.Context) {
	patientID := c.Param("id")
	workspaceID := c.Param("workspace_id")

	cmd := command.TransferCommand{
		ID:         patientID,
		NewParent:  workspaceID,
		ParentType: vobj.EntityTypeWorkspace.String(),
	}

	err := ph.PUseCase.Transfer(c.Request.Context(), cmd)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.NoContent(c)
}

// TransferMany godoc
// @Summary Batch transfer patients
// @Tags Patients
// @Accept json
// @Produce json
// @Param patient_ids query []string true "Patient IDs"
// @Param workspace_id path string true "Target Workspace ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/transfer-many/{workspace_id} [put]
func (ph *PatientHandler) TransferMany(c *gin.Context) {
	ids := c.QueryArray("patient_ids")
	if len(ids) == 0 {
		ph.HandleError(c, errors.NewValidationError("patient_ids parameter is required", nil))
		return
	}

	workspaceID := c.Param("workspace_id")

	cmd := command.TransferManyCommand{
		IDs:        ids,
		NewParent:  workspaceID,
		ParentType: vobj.EntityTypeWorkspace.String(),
	}

	if err := ph.PUseCase.TransferMany(c.Request.Context(), cmd); err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.NoContent(c)
}

// SoftDelete godoc
// @Summary Soft delete patient by ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param id path string true "Patient ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/{id}/soft-delete [delete]
func (ph *PatientHandler) SoftDelete(c *gin.Context) {
	id := c.Param("id")

	err := ph.PQuery.SoftDelete(c.Request.Context(), id)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.NoContent(c)
}

// SoftDeleteMany godoc
// @Summary Batch soft delete patients
// @Tags Patients
// @Accept json
// @Produce json
// @Param ids query []string true "Patient IDs"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/soft-delete-many [delete]
func (ph *PatientHandler) SoftDeleteMany(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		ph.HandleError(c, errors.NewValidationError("ids parameter is required", nil))
		return
	}

	err := ph.PQuery.SoftDeleteMany(c.Request.Context(), ids)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.NoContent(c)
}
