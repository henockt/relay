package api

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/henockt/relay/internal/models"
)

// POST /api/aliases
func (s *Server) handleCreateAlias(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var body struct {
		Label string `json:"label"`
	}
	_ = c.ShouldBindJSON(&body) // label is optional

	alias := &models.Alias{
		UserID:  userID,
		Address: generateAddress(s.cfg.SMTPDomain),
		Label:   body.Label,
		Enabled: true,
	}

	if err := s.aliasStore.Create(alias); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create alias"})
		return
	}

	c.JSON(http.StatusCreated, alias)
}

// GET /api/aliases
func (s *Server) handleListAliases(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	aliases, err := s.aliasStore.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch aliases"})
		return
	}

	c.JSON(http.StatusOK, aliases)
}

// PATCH /api/aliases/:id
func (s *Server) handleUpdateAlias(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	alias, err := s.aliasStore.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
		return
	}

	if alias.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var body struct {
		Label   *string `json:"label"`
		Enabled *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Label != nil {
		alias.Label = *body.Label
	}
	if body.Enabled != nil {
		alias.Enabled = *body.Enabled
	}

	if err := s.aliasStore.Update(alias); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update alias"})
		return
	}

	c.JSON(http.StatusOK, alias)
}

// DELETE /api/aliases/:id
func (s *Server) handleDeleteAlias(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	alias, err := s.aliasStore.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
		return
	}

	if alias.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := s.aliasStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete alias"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// generateAddress creates a random 8-character local part, e.g. "x7k9m2ab@domain"
func generateAddress(domain string) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, 8)
	rand.Read(buf) //nolint:errcheck
	for i, b := range buf {
		buf[i] = charset[int(b) % len(charset)]
	}
	return fmt.Sprintf("%s@%s", string(buf), domain)
}