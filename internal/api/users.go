package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /api/users/me
func (s *Server) handleGetMe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DELETE /api/users/me
func (s *Server) handleDeleteMe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	if err := s.userStore.Delete(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete account"})
		return
	}
	c.Status(http.StatusNoContent)
}
