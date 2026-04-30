package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parseID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func bind[T any](c *gin.Context, dst *T) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func respondError(c *gin.Context, err error, fallbackStatus int, notFoundErrs ...error) {
	for _, notFoundErr := range notFoundErrs {
		if errors.Is(err, notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(fallbackStatus, gin.H{"error": err.Error()})
}
