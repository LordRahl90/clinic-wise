package server

import (
	"context"
	"net/http"

	"clinic-wise/db/models"
	"clinic-wise/internal/services/notes"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type NotesService interface {
	Create(ctx context.Context, req *notes.CreateNoteRequest) (*notes.Response, error)
	Update(ctx context.Context, userID, noteID ulid.ULID, content string) error
	GetAppointmentNotes(ctx context.Context, userID, appointmentID ulid.ULID) ([]notes.Response, error)
	StartDictation(ctx context.Context, conn *websocket.Conn, req *notes.StartDictationRequest) error
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

// createNote godoc
//
//	@Summary		Create a note
//	@Description	Creates a clinical note for an appointment. Requires doctor role.
//	@Tags			Notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		swaggerCreateNoteRequest	true	"Create note payload"
//	@Success		200		{object}	swaggerNoteResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Router			/notes [post]
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

// updateNote godoc
//
//	@Summary		Update a note
//	@Description	Updates the content of an existing note. Requires doctor role.
//	@Tags			Notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		updateNoteRequest	true	"Update note payload"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Router			/notes [patch]
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

// getNote godoc
//
//	@Summary		Get a note
//	@Description	Returns a single note by ID. Requires doctor or patient role and ownership.
//	@Tags			Notes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Note ID (ULID)"
//	@Success		200	{object}	swaggerNoteResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/notes/{id} [get]
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

// getAppointmentNotes godoc
//
//	@Summary		List notes for an appointment
//	@Description	Returns all notes for a given appointment. Requires doctor or patient role and ownership.
//	@Tags			Notes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Appointment ID (ULID)"
//	@Success		200	{array}		swaggerNoteResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/notes/appointment/{id} [get]
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

// startDictation godoc
//
//	@Summary		Start voice dictation (WebSocket)
//	@Description	Upgrades the connection to a WebSocket for real-time voice dictation. Requires doctor role.
//	@Tags			Notes
//	@Security		BearerAuth
//	@Param			body	body		swaggerStartDictationRequest	true	"Dictation request payload"
//	@Success		101	{string}	string	"Switching Protocols"
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Router			/notes/dictation [get]
func (s *Server) startDictation(c *gin.Context) {
	// verify userID
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if user.Role != models.Doctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only doctors can start dictation"})
		return
	}

	var req *notes.StartDictationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DoctorID = user.ID

	w := c.Writer
	r := c.Request

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		httpError(c, err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	err = s.noteService.StartDictation(c.Request.Context(), conn, req)
	if err != nil {
		httpError(c, err)
		return
	}
}
