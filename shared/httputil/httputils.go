// Package httputil provides small handler helpers shared across services.
// It covers the three patterns that repeat in every gin handler:
//   - parsing a UUID path param
//   - binding a JSON request body
//   - mapping a service error to the right HTTP status
package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParseID parses a UUID from the named path parameter.
// On failure it writes a 400 response and returns false — callers should
// return immediately when ok is false.
//
//	id, ok := httputil.ParseID(c, "id")
//	if !ok { return }
func ParseID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

// Bind decodes the JSON request body into dst.
// On failure it writes a 400 response and returns false.
//
//	var req dtos.CreateOrderRequest
//	if !httputil.Bind(c, &req) { return }
func Bind[T any](c *gin.Context, dst *T) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

// RespondError maps a service error to an HTTP response.
// notFoundErrs lists sentinel errors that should produce a 404;
// everything else produces the provided fallbackStatus (typically 400 or 500).
//
//	if err != nil {
//	    httputil.RespondError(c, err, http.StatusInternalServerError,
//	        services.ErrOrderNotFound, services.ErrProductNotFound)
//	    return
//	}
func RespondError(c *gin.Context, err error, fallbackStatus int, notFoundErrs ...error) {
	for _, nfe := range notFoundErrs {
		if errors.Is(err, nfe) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(fallbackStatus, gin.H{"error": err.Error()})
}
