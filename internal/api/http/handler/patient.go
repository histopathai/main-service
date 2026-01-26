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

// CreateNewPatient godoc
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
func (ph *PatientHandler) CreateNewPatient(c *gin.Context) {

	creator_id, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ph.HandleError(c, err)
		return
	}
	var req request.CreatePatientRequest

	if err := c.ShouldBind(&req); err != nil {
		ph.HandleError(c, err)
		return
	}

	// DTO -> Command
	cmd := command.CreatePatientCommand{
		CreateEntityCommand: command.CreateEntityCommand{
			Name:       req.Name,
			EntityType: vobj.EntityTypePatient.String(),
			CreatorID:  creator_id,
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

	ph.Response.Success(c, http.StatusCreated, response.NewPatientResponse(patient))

}

// Get [get] godoc
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

// GetByParentID [get] godoc
// @Summary Get patients by workspace ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body request.ListRequest false "List request"
// @Success 200 {object} response.PatientListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces/{parent_id}/patients [get]
func (ph *PatientHandler) GetByParentID(c *gin.Context) {
	var req request.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = request.ListRequest{}
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

	result, err := ph.PQuery.GetByParentID(c.Request.Context(), spec, c.Param("parent_id"))
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	paginationResp := response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	// Service Output -> DTO
	patientResponses := make([]response.PatientResponse, len(result.Data))
	for i, patient := range result.Data {
		patientResponses[i] = *response.NewPatientResponse(patient)
	}

	ph.Response.SuccessList(c, patientResponses, &paginationResp)

}

// List [post] godoc
// @Summary      List patients
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Param        request body request.ListRequest
// @Success      200 {object} response.PatientListResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /patients/list [post]
func (ph *PatientHandler) List(c *gin.Context) {
	var req request.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.HandleError(c, errors.NewValidationError("invalid request payload",
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

	result, err := ph.PQuery.List(c.Request.Context(), spec)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	paginationResp := response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}
	patientResponses := make([]response.PatientResponse, len(result.Data))
	for i, patient := range result.Data {
		patientResponses[i] = *response.NewPatientResponse(patient)
	}

	ph.Response.SuccessList(c, patientResponses, &paginationResp)

}

// Count [post]	godoc
// @Summary Count patients
// @Tags Patients
// @Accept json
// @Produce json
// @Param request body request.ListRequest
// @Success 200 {object} response.CountResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /patients/count [post]
func (ph *PatientHandler) Count(c *gin.Context) {
	var req request.ListRequest
	// Optional bind
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.HandleError(c, errors.NewValidationError("invalid request payload",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ph.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	count, err := ph.PQuery.Count(c.Request.Context(), spec)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	countResp := response.CountResponse{
		Count: count,
	}

	ph.Response.Success(c, http.StatusOK, countResp)
}

// Update [put] godoc
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

	var req request.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.HandleError(c, errors.NewValidationError("invalid request payload",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	// DTO -> Command

	cmd := command.UpdatePatientCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{
			ID:        c.Param("id"),
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

	err := ph.PQuery.SoftDeleteMany(c.Request.Context(), ids)
	if err != nil {
		ph.HandleError(c, err)
		return
	}

	ph.Response.NoContent(c)
}
