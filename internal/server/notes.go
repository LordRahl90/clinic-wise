package server

import (
	"context"
	"net/http"

	"clinic-wise/internal/services/notes"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type NotesService interface {
	Create(ctx context.Context, req *notes.CreateNoteRequest) (*notes.Response, error)
	Update(ctx context.Context, userID, noteID ulid.ULID, content string) error
	GetAppointmentNotes(ctx context.Context, userID, appointmentID ulid.ULID) ([]notes.Response, error)
}

func (s *Server) noteRoutes() {
	noteGroup := s.router.Group("/notes")
	noteGroup.Use(s.authMiddleware.Middleware())
	{
		noteGroup.POST("", s.createNote)
		noteGroup.PATCH("", s.updateNote)
		noteGroup.GET("/:id", s.getNote)
		noteGroup.GET("/appointment/:id", s.getAppointmentNotes)
	}
}

func (s *Server) createNote(c *gin.Context) {
	var req notes.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.noteService.Create(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) updateNote(c *gin.Context) {

}

func (s *Server) getNote(c *gin.Context) {
}

func (s *Server) getAppointmentNotes(c *gin.Context) {
}
