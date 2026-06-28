package server

import (
	"context"
	"net/http"
	"strconv"

	"clinic-wise/db/models"
	"clinic-wise/internal/server/middlewares"
	"clinic-wise/internal/services/appointments"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type appointmentService interface {
	Create(ctx context.Context, req *appointments.CreateAppointmentRequest) (*appointments.Response, error)
	Complete(ctx context.Context, userID, id ulid.ULID) (*appointments.Response, error)
	Find(ctx context.Context, userID, id ulid.ULID) (*appointments.Response, error)
	FindAppointments(ctx context.Context, hospitalID ulid.ULID) ([]appointments.Response, error)
	FindAppointmentByUser(ctx context.Context, userID ulid.ULID, page, limit int) ([]appointments.Response, error)
}

func (s *Server) appointmentRoutes() {
	s.router.GET("/hospitals/:id/appointments", s.authMiddleware.Middleware(), s.findHospitalAppointments)
	appointment := s.router.Group("/appointments")
	appointment.Use(s.authMiddleware.Middleware())
	{
		appointment.POST("", s.createAppointment)

		appointment.GET("/user", s.findAppointmentsByUser)
		appointment.GET("/:id", s.findAppointment)
		appointment.PATCH("/:id/complete", s.completeAppointment)
	}
}

// createAppointment godoc
//
//	@Summary		Create an appointment
//	@Description	Creates a new appointment. Requires authentication.
//	@Tags			Appointments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		swaggerCreateAppointmentRequest	true	"Create appointment payload"
//	@Success		200		{object}	swaggerAppointmentResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/appointments [post]
func (s *Server) createAppointment(c *gin.Context) {
	var request *appointments.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middlewares.ExtractUserInfo(c, s.config.SigningSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	request.ActorID = userID.ID.String()

	res, err := s.appointmentService.Create(c.Request.Context(), request)
	if err != nil {
		httpError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// findHospitalAppointments godoc
//
//	@Summary		List appointments for a hospital
//	@Description	Returns all appointments for the given hospital. Requires authentication.
//	@Tags			Appointments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			hospitalId	path		string	true	"Hospital ID (ULID)"
//	@Success		200			{array}		swaggerAppointmentResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Router			/hospitals/{hospitalId}/appointments [get]
func (s *Server) findHospitalAppointments(c *gin.Context) {
	hospitalIDStr := c.Param("id")
	if hospitalIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hospitalId is required"})
		return
	}

	hospitalID, err := ulid.ParseStrict(hospitalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospitalId"})
		return
	}

	res, err := s.appointmentService.FindAppointments(c.Request.Context(), hospitalID)
	if err != nil {
		httpError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// findAppointmentsByUser godoc
//
//	@Summary		List appointments for the current user
//	@Description	Returns paginated appointments belonging to the authenticated user.
//	@Tags			Appointments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	true	"Page number (1-based)"
//	@Param			limit	query		int	true	"Page size"
//	@Success		200		{array}		swaggerAppointmentResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/appointments/user [get]
func (s *Server) findAppointmentsByUser(c *gin.Context) {
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
		httpError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// findAppointment godoc
//
//	@Summary		Get an appointment
//	@Description	Returns a single appointment by ID. Requires authentication.
//	@Tags			Appointments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Appointment ID (ULID)"
//	@Success		200	{object}	swaggerAppointmentResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Router			/appointments/{id} [get]
func (s *Server) findAppointment(c *gin.Context) {
	id, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, err := middlewares.ExtractUserInfo(c, s.config.SigningSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	res, err := s.appointmentService.Find(c.Request.Context(), userID.ID, id)
	if err != nil {
		httpError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// completeAppointment godoc
//
//	@Summary		Complete an appointment
//	@Description	Marks an appointment as completed. Requires doctor role.
//	@Tags			Appointments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Appointment ID (ULID)"
//	@Success		200	{object}	swaggerAppointmentResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Router			/appointments/{id}/complete [patch]
func (s *Server) completeAppointment(c *gin.Context) {
	user, err := middlewares.ExtractUserInfo(c, s.config.SigningSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if user.Role != models.Doctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors can complete appointments"})
		return
	}

	id, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.appointmentService.Complete(c.Request.Context(), user.ID, id)
	if err != nil {
		httpError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
