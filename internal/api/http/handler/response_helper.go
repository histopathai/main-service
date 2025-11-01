package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/response"
)

type ResponseHelper struct{}

func (rh *ResponseHelper) Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, response.DataResponse[interface{}]{
		Data: data,
	})
}

func (rh *ResponseHelper) SuccessList(c *gin.Context, data interface{}, pagination *response.PaginationResponse) {
	listData, _ := data.([]interface{})
	c.JSON(http.StatusOK, response.ListResponse[interface{}]{
		Data:       listData,
		Pagination: pagination,
	})
}

func (rh *ResponseHelper) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (rh *ResponseHelper) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, response.DataResponse[interface{}]{
		Data: data,
	})
}
