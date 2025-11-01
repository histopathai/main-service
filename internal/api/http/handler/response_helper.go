package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/response"
)

type ResponseHelper struct{}

func (rh *ResponseHelper) Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{
		"data": data,
	})
}

func (rh *ResponseHelper) SuccessList(c *gin.Context, data interface{}, pagination *response.PaginationResponse) {
	response := gin.H{
		"data": data,
	}

	if pagination != nil {
		response["pagination"] = pagination
	}

	c.JSON(http.StatusOK, response)
}
func (rh *ResponseHelper) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (rh *ResponseHelper) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"data": data,
	})
}

func (rh *ResponseHelper) Error(c *gin.Context, statusCode int, errType string, message string, details map[string]interface{}) {
	c.JSON(statusCode, response.ErrorResponse{
		ErrorType: errType,
		Message:   message,
		Details:   details,
	})
}
