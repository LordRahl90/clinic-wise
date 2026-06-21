package server

import (
	"context"
	"net/http"

	"clinic-wise/db/models"
	"clinic-wise/internal/server/middlewares"
	"clinic-wise/internal/services/hospital"

	"github.com/gin-gonic/gin"
)

type HospitalsService interface {
	Create(ctx context.Context, req *hospital.CreateHospitalRequest) (*hospital.Response, error)
}

func (s *Server) hospitalRoutes() {
	h := s.router.Group("/hospitals")
	{
		h.POST("", s.authMiddleware.Middleware(), s.createHospital)
	}
}

func (s *Server) createHospital(c *gin.Context) {
	var req *hospital.CreateHospitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := middlewares.ExtractUserInfo(c, s.config.SigningSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != models.Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can create hospitals"})
		return
	}

	req.UserID = user.ID
	res, err := s.hospitalService.Create(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
