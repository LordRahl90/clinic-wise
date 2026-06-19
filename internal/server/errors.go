package server

import (
	"errors"
	"net/http"

	audittrailservice "clinic-wise/internal/services/audittrail"
	prescriptionsservice "clinic-wise/internal/services/prescriptions"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// httpError writes the most specific status code for a service error.
//
//   - gorm.ErrRecordNotFound → 404 Not Found
//   - audittrailservice.ErrForbidden → 403 Forbidden
//   - prescriptionsservice.ErrPrescriptionExpired/ErrPrescriptionUnavailable → 400 Bad Request
//   - everything else → 500 Internal Server Error
func httpError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, audittrailservice.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, prescriptionsservice.ErrPrescriptionExpired),
		errors.Is(err, prescriptionsservice.ErrPrescriptionUnavailable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

