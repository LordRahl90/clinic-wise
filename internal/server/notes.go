package server

import (
	"context"
	"net/http"

	"clinic-wise/db/models"
	"clinic-wise/internal/services/notes"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type NotesService interface {
	Create(ctx context.Context, req *notes.CreateNoteRequest) (*notes.Response, error)
	Update(ctx context.Context, userID, noteID ulid.ULID, content string) error
	GetAppointmentNotes(ctx context.Context, userID, appointmentID ulid.ULID) ([]notes.Response, error)
}

type updateNoteRequest struct {
	NoteID  string `json:"note_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (s *Server) noteRoutes() {
	noteGroup := s.router.Group("/notes")
	noteGroup.Use(s.authMiddleware.Middleware())
	{
		noteGroup.POST("", s.createNote)
		noteGroup.PATCH("", s.updateNote)
		noteGroup.GET("/:id", s.getNote)
		noteGroup.GET("/appointment/:id", s.getAppointmentNotes)
		noteGroup.GET("/dictation", s.startDictation)
	}
}

func (s *Server) createNote(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors can create notes"})
		return
	}

	var req notes.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DoctorID = user.ID

	res, err := s.noteService.Create(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) updateNote(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors can update notes"})
		return
	}

	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	noteID, err := ulid.ParseStrict(req.NoteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note_id"})
		return
	}

	if err := s.noteService.Update(c.Request.Context(), user.ID, noteID, req.Content); err != nil {
		httpError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) getNote(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor && user.Role != models.Patient {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors and patients can view notes"})
		return
	}

	noteID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var note models.Note
	if err := s.config.DB.WithContext(c.Request.Context()).Where("id = ?", noteID).First(&note).Error; err != nil {
		httpError(c, err)
		return
	}

	if note.DoctorID != user.ID && note.PatientID != user.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, notes.FromModel(&note))
}

func (s *Server) getAppointmentNotes(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor && user.Role != models.Patient {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors and patients can view notes"})
		return
	}

	appointmentID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var appointment models.Appointment
	if err := s.config.DB.WithContext(c.Request.Context()).Where("id = ?", appointmentID).First(&appointment).Error; err != nil {
		httpError(c, err)
		return
	}
	if appointment.DoctorID != user.ID && appointment.PatientID != user.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	res, err := s.noteService.GetAppointmentNotes(c.Request.Context(), user.ID, appointmentID)
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) startDictation(c *gin.Context) {

}
