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
