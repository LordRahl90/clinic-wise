package server

import (
	"context"
	"net/http"

	"clinic-wise/db/models"
	"clinic-wise/internal/server/middlewares"
	"clinic-wise/internal/services/hospital"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type HospitalsService interface {
	Create(ctx context.Context, req *hospital.CreateHospitalRequest) (*hospital.CreateHospitalResponse, error)
	Stats(ctx context.Context, hospitalID ulid.ULID) (*hospital.StatsResponse, error)
}

func (s *Server) hospitalRoutes() {
	h := s.router.Group("/hospitals")
	{
		h.POST("", s.authMiddleware.Middleware(), s.createHospital)
		h.GET("/:id/stats", s.authMiddleware.Middleware(), s.hospitalStats)
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

func (s *Server) hospitalStats(c *gin.Context) {
	user, err := middlewares.ExtractUserInfo(c, s.config.SigningSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if user.Role != models.Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can view hospital stats"})
		return
	}

	hospitalID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.hospitalService.Stats(c.Request.Context(), hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
