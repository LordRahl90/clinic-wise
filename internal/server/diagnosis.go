package server

import (
	"context"
	"net/http"

	"clinic-wise/db/models"
	diagnosisservice "clinic-wise/internal/services/diagnosis"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type DiagnosisService interface {
	Create(ctx context.Context, req *diagnosisservice.CreateDiagnosisRequest) (*diagnosisservice.Response, error)
	Find(ctx context.Context, userID, diagnosisID ulid.ULID) (*diagnosisservice.Response, error)
	Dismiss(ctx context.Context, doctorID, diagnosisID ulid.ULID) (*diagnosisservice.Response, error)
	FindByUser(ctx context.Context, userID ulid.ULID) ([]diagnosisservice.Response, error)
}

func (s *Server) diagnosisRoutes() {
	diagnosisGroup := s.router.Group("/diagnoses")
	diagnosisGroup.Use(s.authMiddleware.Middleware())
	{
		diagnosisGroup.POST("", s.createDiagnosis)
		diagnosisGroup.GET("", s.listDiagnoses)
		diagnosisGroup.GET("/:id", s.getDiagnosis)
		diagnosisGroup.PATCH("/:id/dismiss", s.dismissDiagnosis)
	}
}

func (s *Server) createDiagnosis(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors can create diagnoses"})
		return
	}

	var req diagnosisservice.CreateDiagnosisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DoctorID = user.ID

	res, err := s.diagnosisService.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) getDiagnosis(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor && user.Role != models.Patient {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors and patients can view diagnoses"})
		return
	}

	diagnosisID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.diagnosisService.Find(c.Request.Context(), user.ID, diagnosisID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) listDiagnoses(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor && user.Role != models.Patient {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors and patients can view diagnoses"})
		return
	}

	res, err := s.diagnosisService.FindByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) dismissDiagnosis(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors can dismiss diagnoses"})
		return
	}

	diagnosisID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.diagnosisService.Dismiss(c.Request.Context(), user.ID, diagnosisID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}



