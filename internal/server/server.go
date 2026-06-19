package server

import (
	"clinic-wise/internal/server/middlewares"
	"clinic-wise/internal/services/appointments"
	"clinic-wise/internal/services/hospital"
	"clinic-wise/internal/services/integrations/queue"
	"clinic-wise/internal/services/notes"
	"log/slog"

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

	appointmentService appointmentService
	hospitalService    HospitalsService
	noteService        NotesService
}

func New(config *Config) *Server {
	router := gin.Default()

	s := &Server{
		router:         router,
		config:         config,
		authMiddleware: middlewares.NewAuthMiddleware(config.SigningSecret),

		hospitalService:    hospital.New(config.DB),
		appointmentService: appointments.New(config.DB),
		noteService:        notes.New(config.DB, config.Writer),
	}

	// register routes
	s.hospitalRoutes()
	s.appointmentRoutes()
	s.noteRoutes()

	return s
}

func (s *Server) Run() error {
	slog.Info("Starting server on port", "port", s.config.Port)
	return s.router.Run(":" + s.config.Port)
}
