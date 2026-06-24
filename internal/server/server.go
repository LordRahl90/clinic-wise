package server

import (
	"log/slog"

	"clinic-wise/internal/server/middlewares"
	"clinic-wise/internal/services/appointments"
	"clinic-wise/internal/services/audittrail"
	authservice "clinic-wise/internal/services/auth"
	"clinic-wise/internal/services/diagnosis"
	"clinic-wise/internal/services/hospital"
	"clinic-wise/internal/services/integrations/queue"
	"clinic-wise/internal/services/notes"
	"clinic-wise/internal/services/prescriptions"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Config struct {
	DB            *gorm.DB
	Port          string
	SigningSecret string
	Writer        *queue.Service
}

type Server struct {
	router *gin.Engine
	config *Config

	authMiddleware *middlewares.AuthMiddleware

	authService         AuthService
	appointmentService  appointmentService
	auditTrailService   AuditTrailService
	diagnosisService    DiagnosisService
	hospitalService     HospitalsService
	noteService         NotesService
	prescriptionService PrescriptionService
}

func New(config *Config) *Server {
	router := gin.Default()
	noteWriter := notes.NewNoopWriter()
	if config.Writer != nil {
		noteWriter = config.Writer
	}
	prescriptionWriter := prescriptions.NewNoopEventWriter()
	if config.Writer != nil {
		prescriptionWriter = config.Writer
	}

	s := &Server{
		router:         router,
		config:         config,
		authMiddleware: middlewares.NewAuthMiddleware(config.SigningSecret),
		authService:    authservice.New(config.DB, config.SigningSecret),

		hospitalService:     hospital.New(config.DB),
		appointmentService:  appointments.New(config.DB),
		auditTrailService:   audittrail.New(config.DB),
		diagnosisService:    diagnosis.New(config.DB),
		noteService:         notes.New(config.DB, noteWriter),
		prescriptionService: prescriptions.New(config.DB, prescriptionWriter),
	}

	// register routes
	s.hospitalRoutes()
	s.authRoutes()
	s.appointmentRoutes()
	s.auditTrailRoutes()
	s.diagnosisRoutes()
	s.noteRoutes()
	s.prescriptionRoutes()

	return s
}

func (s *Server) Run() error {
	slog.Info("Starting server on port", "port", s.config.Port)
	return s.router.Run(":" + s.config.Port)
}
