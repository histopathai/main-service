package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"
)

func parsePagination(c *gin.Context) (repository.Pagination, error) {
	const (
		defaultLimit  = 10
		maxLimit      = 100
		defaultOffset = 0
	)

	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return repository.Pagination{}, apperrors.NewValidationError("invalid limit parameter", nil)
	}

	offsetStr := c.DefaultQuery("offset", strconv.Itoa(defaultOffset))
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return repository.Pagination{}, apperrors.NewValidationError("invalid offset parameter", nil)
	}

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = defaultOffset
	}

	return repository.Pagination{
		Limit:  limit,
		Offset: offset,
	}, nil
}
