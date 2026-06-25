package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	audittrailservice "clinic-wise/internal/services/audittrail"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type AuditTrailService interface {
	FindByAppointment(ctx context.Context, userID, appointmentID ulid.ULID, filter audittrailservice.FilterQuery) ([]audittrailservice.Response, error)
	FindByEntity(ctx context.Context, userID ulid.ULID, entityType, entityID string, filter audittrailservice.FilterQuery) ([]audittrailservice.Response, error)
}

func (s *Server) auditTrailRoutes() {
	group := s.router.Group("/audit-trails")
	group.Use(s.authMiddleware.Middleware())
	{
		group.GET("/appointment/:id", s.listAppointmentAuditTrail)
		group.GET("/entity/:type/:id", s.listEntityAuditTrail)
	}
}

func parseFilterQuery(c *gin.Context) audittrailservice.FilterQuery {
	f := audittrailservice.FilterQuery{}

	f.Action = c.Query("action")

	if raw := c.Query("actor_id"); raw != "" {
		if id, err := ulid.ParseStrict(raw); err == nil {
			f.ActorID = id
		}
	}
	if raw := c.Query("from"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			f.From = t
		}
	}
	if raw := c.Query("to"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			f.To = t
		}
	}
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		f.Page = p
	}
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		f.Limit = l
	}
	return f
}

// listAppointmentAuditTrail godoc
//
//	@Summary		List audit trail for an appointment
//	@Description	Returns audit log entries for a given appointment. Requires authentication and ownership.
//	@Tags			Audit Trails
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Appointment ID (ULID)"
//	@Param			action		query		string	false	"Filter by action"
//	@Param			actor_id	query		string	false	"Filter by actor ID (ULID)"
//	@Param			from		query		string	false	"Filter from date (RFC3339)"
//	@Param			to			query		string	false	"Filter to date (RFC3339)"
//	@Param			page		query		int		false	"Page number (1-based)"
//	@Param			limit		query		int		false	"Page size (max 200, default 50)"
//	@Success		200			{array}		swaggerAuditTrailResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Router			/audit-trails/appointment/{id} [get]
func (s *Server) listAppointmentAuditTrail(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	appointmentID, err := ulid.ParseStrict(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := s.auditTrailService.FindByAppointment(c.Request.Context(), user.ID, appointmentID, parseFilterQuery(c))
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// listEntityAuditTrail godoc
//
//	@Summary		List audit trail for an entity
//	@Description	Returns audit log entries for a given entity type and ID. Requires authentication.
//	@Tags			Audit Trails
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type		path		string	true	"Entity type (e.g. user, appointment)"
//	@Param			id			path		string	true	"Entity ID"
//	@Param			action		query		string	false	"Filter by action"
//	@Param			actor_id	query		string	false	"Filter by actor ID (ULID)"
//	@Param			from		query		string	false	"Filter from date (RFC3339)"
//	@Param			to			query		string	false	"Filter to date (RFC3339)"
//	@Param			page		query		int		false	"Page number (1-based)"
//	@Param			limit		query		int		false	"Page size (max 200, default 50)"
//	@Success		200			{array}		swaggerAuditTrailResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Router			/audit-trails/entity/{type}/{id} [get]
func (s *Server) listEntityAuditTrail(c *gin.Context) {
	user := currentUserInfo(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	entityType := c.Param("type")
	entityID := c.Param("id")
	if entityType == "" || entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity type and id are required"})
		return
	}

	res, err := s.auditTrailService.FindByEntity(c.Request.Context(), user.ID, entityType, entityID, parseFilterQuery(c))
	if err != nil {
		httpError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
