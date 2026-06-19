package server

import (
	"context"
	"net/http"

	"clinic-wise/db/models"
	prescriptionsservice "clinic-wise/internal/services/prescriptions"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type PrescriptionService interface {
	Create(ctx context.Context, req *prescriptionsservice.CreatePrescriptionRequest) (*prescriptionsservice.Response, error)
	Dispatch(ctx context.Context, pharmacistID, prescriptionID ulid.ULID) (*prescriptionsservice.Response, error)
	Find(ctx context.Context, userID, prescriptionID ulid.ULID) (*prescriptionsservice.Response, error)
	FindByAppointment(ctx context.Context, userID, appointmentID ulid.ULID) ([]prescriptionsservice.Response, error)
}

func (s *Server) prescriptionRoutes() {
	prescriptionsGroup := s.router.Group("/prescriptions")
	prescriptionsGroup.Use(s.authMiddleware.Middleware())
	{
		prescriptionsGroup.POST("", s.createPrescription)
		prescriptionsGroup.PATCH("/:id/dispatch", s.dispatchPrescription)
		prescriptionsGroup.GET("/:id", s.getPrescription)
		prescriptionsGroup.GET("/appointment/:id", s.getAppointmentPrescriptions)
	}
}

func (s *Server) createPrescription(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors can create prescriptions"})
		return
	}

	var req prescriptionsservice.CreatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DoctorID = user.ID.String()

	res, err := s.prescriptionService.Create(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) dispatchPrescription(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Pharmacist {
		c.JSON(http.StatusForbidden, gin.H{"error": "only pharmacists can dispatch prescriptions"})
		return
	}

	prescriptionID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.prescriptionService.Dispatch(c.Request.Context(), user.ID, prescriptionID)
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) getPrescription(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor && user.Role != models.Patient {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors and patients can view prescriptions"})
		return
	}

	prescriptionID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.prescriptionService.Find(c.Request.Context(), user.ID, prescriptionID)
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) getAppointmentPrescriptions(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor && user.Role != models.Patient {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors and patients can view prescriptions"})
		return
	}

	appointmentID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.prescriptionService.FindByAppointment(c.Request.Context(), user.ID, appointmentID)
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
