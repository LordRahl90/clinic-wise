package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"clinic-wise/internal/server/middlewares"
	"clinic-wise/internal/services/appointments"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type appointmentService interface {
	Create(ctx context.Context, req *appointments.CreateAppointmentRequest) (*appointments.Response, error)
	Find(ctx context.Context, userID, id ulid.ULID) (*appointments.Response, error)
	FindAppointments(ctx context.Context, hospitalID ulid.ULID) ([]appointments.Response, error)
	FindAppointmentByUser(ctx context.Context, userID ulid.ULID, page, limit int) ([]appointments.Response, error)
}

func (s *Server) appointmentRoutes() {
	appointment := s.router.Group("/appointments")
	appointment.Use(s.authMiddleware.Middleware())
	{
		appointment.POST("", s.createAppointment)
		appointment.GET("", s.findAppointments)
		appointment.GET("/user", s.findAppointmentsByUser)
		appointment.GET("/:id", s.findAppointment)
	}
}

func (s *Server) createAppointment(c *gin.Context) {
	var request *appointments.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.appointmentService.Create(c.Request.Context(), request)
	if err != nil {
		// we should check for ulid format errors as well
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) findAppointments(c *gin.Context) {
	hospitalIDStr := c.Query("hospital_id")
	if hospitalIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hospital_id is required"})
		return
	}

	hospitalID, err := ulid.ParseStrict(hospitalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospital_id"})
		return
	}

	res, err := s.appointmentService.FindAppointments(c.Request.Context(), hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) findAppointmentsByUser(c *gin.Context) {
	// extract userID from middleware
	userID, err := middlewares.ExtractUserInfo(c, s.config.SigningSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	res, err := s.appointmentService.FindAppointmentByUser(c.Request.Context(), userID.ID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) findAppointment(c *gin.Context) {
	id, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// extract userID from middleware
	userID, err := middlewares.ExtractUserInfo(c, s.config.SigningSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	res, err := s.appointmentService.Find(c.Request.Context(), userID.ID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
