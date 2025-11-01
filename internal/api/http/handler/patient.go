package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/request"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/response"
	"github.com/histopathai/main-service-refactor/internal/api/http/middleware"
	"github.com/histopathai/main-service-refactor/internal/api/http/validator"
	"github.com/histopathai/main-service-refactor/internal/service"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type PatientHandler struct {
	patientService *service.PatientService
	validator      *validator.RequestValidator
	BaseHandler    // Embed the BaseHandler
}

func NewPatientHandler(patientService *service.PatientService, validator *validator.RequestValidator, logger *slog.Logger) *PatientHandler {
	return &PatientHandler{
		patientService: patientService,
		validator:      validator,
		BaseHandler:    BaseHandler{logger: logger},
	}
}

// CreateNewPatient godoc
// @Summary Create a new patient
// @Description Create a new patient with the provided details
// @Tags Patients
// @Accept json
// @Produce json
// @Param        request body request.CreatePatientRequest true "Patient creation request"
// @Success 201 {object} response.PatientDataResponse "Patient created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /patients [post]

func (ph *PatientHandler) CreateNewPatient(c *gin.Context) {

	creator_id, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ph.handleError(c, err)
		return
	}
	var req request.CreatePatientRequest

	if err := c.ShouldBind(&req); err != nil {
		ph.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	if err := ph.validator.ValidateStruct(&req); err != nil {
		ph.handleError(c, err)
		return
	}

	// DTO -> Service Input
	input := service.CreatePatientInput{
		WorkspaceID: req.WorkspaceID,
		CreatorID:   creator_id,
		Name:        req.Name,
		Age:         req.Age,
		Gender:      req.Gender,
		Race:        req.Race,
		Disease:     req.Disease,
		Subtype:     req.Subtype,
		Grade:       req.Grade,
		History:     req.History,
	}

	patient, err := ph.patientService.CreateNewPatient(c.Request.Context(), input)
	if err != nil {
		ph.handleError(c, err)
		return
	}
	ph.logger.Info("Patient created", slog.String("patient_id", patient.ID))

	// Service Output -> DTO
	patientResp := response.NewPatientResponse(patient)

	ph.response.Created(c, patientResp)
}

// GetPatientByID godoc
// @Summary Get patient by ID
// @Description Retrieve patient details by patient ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Success 200 {object} response.PatientDataResponse "Patient retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid patient ID"
// @Failure 404 {object} response.ErrorResponse "Patient not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /patients/{patient_id} [get]

func (ph *PatientHandler) GetPatientByID(c *gin.Context) {
	patientID := c.Param("patient_id")

	patient, err := ph.patientService.GetPatientByID(c.Request.Context(), patientID)
	if err != nil {
		ph.handleError(c, err)
		return
	}

	ph.logger.Info("Patient retrieved successfully",
		slog.String("patient_id", patient.ID))

	// Service Output -> DTO
	patientResp := response.NewPatientResponse(patient)
	ph.response.Success(c, http.StatusOK, patientResp)
}

// GetPatientsByWorkspaceID godoc
// @Summary Get patients by workspace ID
// @Description Retrieve a list of patients associated with a specific workspace ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param workspace_id path string true "Workspace ID"
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.PatientListResponse "Patients retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid workspace ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /workspaces/{workspace_id}/patients [get]

func (ph *PatientHandler) GetPatientsByWorkspaceID(c *gin.Context) {
	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		ph.handleError(c, errors.NewValidationError("invalid query parameters",
			map[string]interface{}{"error": err.Error()}))
		return
	}
	workspaceID := c.Param("workspace_id")

	pagination := &query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := ph.patientService.GetPatientsByWorkspaceID(c.Request.Context(), workspaceID, pagination)
	if err != nil {
		ph.handleError(c, err)
		return
	}

	ph.logger.Info("Patients retrieved successfully",
		slog.String("workspace_id", workspaceID))

	paginationResp := response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
		Total:   result.Total,
	}

	// Service Output -> DTO
	patientResponses := make([]response.PatientResponse, len(result.Data))
	for i, patient := range result.Data {
		patientResponses[i] = *response.NewPatientResponse(patient)
	}

	ph.response.SuccessList(c, patientResponses, &paginationResp)

}

// UpdatePatientByID godoc
// @Summary Update patient by ID
// @Description Update patient details using their ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Param        request body request.UpdatePatientRequest true "Patient update request"
// @Success 204 "Patient updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "Patient not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /patients/{patient_id} [put]
func (ph *PatientHandler) UpdatePatientByID(c *gin.Context) {
	patientID := c.Param("patient_id")

	var req request.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// DTO -> Service Input
	input := service.UpdatePatientInput{
		Name:    req.Name,
		Age:     req.Age,
		Race:    req.Race,
		Gender:  req.Gender,
		Disease: req.Disease,
		Subtype: req.Subtype,
		Grade:   req.Grade,
		History: req.History,
	}

	err := ph.patientService.UpdatePatient(c.Request.Context(), patientID, input)
	if err != nil {
		ph.handleError(c, err)
		return
	}

	ph.logger.Info("Patient updated successfully",
		slog.String("patient_id", patientID))

	// No content to return
	ph.response.NoContent(c)

}

// ListPatients godoc
// @Summary      List patients
// @Description  Get paginated list of patients
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success      200 {object} response.PatientListResponse "List of patients"
// @Failure      400 {object} response.ErrorResponse "Invalid query parameters"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /patients [get]

func (ph *PatientHandler) ListPatients(c *gin.Context) {
	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		ph.handleError(c, errors.NewValidationError("invalid query parameters",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	pagination := &query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := ph.patientService.ListPatients(c.Request.Context(), pagination)
	if err != nil {
		ph.handleError(c, err)
		return
	}
	ph.logger.Info("Patients listed successfully")

	// Service Output -> DTO
	paginationResp := response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
		Total:   result.Total,
	}
	patientResponses := make([]response.PatientResponse, len(result.Data))
	for i, patient := range result.Data {
		patientResponses[i] = *response.NewPatientResponse(patient)
	}

	ph.response.SuccessList(c, patientResponses, &paginationResp)

}

// DeletePatientByID godoc
// @Summary Delete patient by ID
// @Description Delete a patient using their ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Success 204 "Patient deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid patient ID"
// @Failure 404 {object} response.ErrorResponse "Patient not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /patients/{patient_id} [delete]

func (ph *PatientHandler) DeletePatientByID(c *gin.Context) {
	patientID := c.Param("patient_id")

	err := ph.patientService.DeletePatientByID(c.Request.Context(), patientID)
	if err != nil {
		ph.handleError(c, err)
		return
	}

	ph.logger.Info("Patient deleted successfully",
		slog.String("patient_id", patientID),
	)

	ph.response.NoContent(c)
}

//TransferPatientWorkspace godoc
// @Summary Transfer patient to another workspace
// @Description Transfer a patient to a different workspace using their ID
// @Tags Patients
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Param workspace_id path string true "Target Workspace ID"
// @Success 204 "Patient transferred successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid patient ID or workspace ID"
// @Failure 404 {object} response.ErrorResponse "Patient or workspace not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /patients/{patient_id}/transfer/{workspace_id} [post]

func (ph *PatientHandler) TransferPatientWorkspace(c *gin.Context) {
	patientID := c.Param("patient_id")
	workspaceID := c.Param("workspace_id")

	err := ph.patientService.TransferPatientWorkspace(c.Request.Context(), patientID, workspaceID)
	if err != nil {
		ph.handleError(c, err)
		return
	}

	ph.logger.Info("Patient transferred successfully",
		slog.String("patient_id", patientID),
		slog.String("target_workspace_id", workspaceID),
	)
	ph.response.NoContent(c)
}
