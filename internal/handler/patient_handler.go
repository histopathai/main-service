package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/service"
)

type PatientHandler struct {
	patientService *service.PatientService
	logger         *slog.Logger
	validate       *validator.Validate
}

func NewPatientHandler(repo *repository.MainRepository, logger *slog.Logger) *PatientHandler {
	return &PatientHandler{
		patientService: service.NewPatientService(
			repository.NewPatientRepository(*repo),
			logger),
		logger:   logger,
		validate: validator.New(),
	}
}

func (ph *PatientHandler) RegisterRoutes(rg *gin.RouterGroup) {
	patients := rg.Group("/patients")
	patients.POST("/", ph.CreatePatient)
	patients.GET("/:patient_id", ph.GetPatient)
	patients.GET("/", ph.GetAllPatients)
	patients.GET("/workspace/:workspace_id", ph.GetPatientsByWorkspaceID)
	patients.PUT("/:patient_id/move/workspace/:workspace_id", ph.MovePatientToWorkspace)
}

func (ph *PatientHandler) CreatePatient(c *gin.Context) {
	var req CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := ph.patientService.CreatePatient(c.Request.Context(), req.ToModel())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (ph *PatientHandler) GetPatient(c *gin.Context) {
	patientID := c.Param("patient_id")
	resp, err := ph.patientService.GetPatient(c.Request.Context(), patientID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (ph *PatientHandler) GetAllPatients(c *gin.Context) {

	pagination, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid pagination parameters",
			"details": err.Error(),
		})
		return
	}

	resp, err := ph.patientService.GetAllPatients(c.Request.Context(), pagination)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (ph *PatientHandler) GetPatientsByWorkspaceID(c *gin.Context) {
	workspaceID := c.Param("workspace_id")

	pagination, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid pagination parameters",
			"details": err.Error(),
		})
		return
	}

	resp, err := ph.patientService.GetPatientsByWorkspaceID(c.Request.Context(), workspaceID, pagination)
	if err != nil {
		handleError(c, err)
		ph.logger.Error("failed to list patients by workspace ID", "error", err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (ph *PatientHandler) MovePatientToWorkspace(c *gin.Context) {
	workspace_id := c.Param("workspace_id")

	patient_id := c.Param("patient_id")

	err := ph.patientService.MovePatientToWorkspace(c.Request.Context(), patient_id, workspace_id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "patient moved successfully",
	})
}
