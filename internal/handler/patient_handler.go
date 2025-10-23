package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/service"
)

type PatientHandler struct {
	patientService *service.PatientService
}

func NewPatientHandler(repo *repository.MainRepository, logger *slog.Logger) *PatientHandler {
	return &PatientHandler{
		patientService: service.NewPatientService(
			repository.NewPatientRepository(repo),
			logger),
	}
}

func (ph *PatientHandler) CreatePatient(c *gin.Context) {
	var req service.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}
	resp, err := ph.patientService.CreatePatient(c.Request.Context(), &req)
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
