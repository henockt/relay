package api

import (
	"log"
	"net/http"

	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
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
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete account"})
		return
	}
	if err := s.userStore.Delete(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete account"})
		return
	}
	// also remove the Clerk identity so it doesn't linger as an orphan.
	// relay data is already gone, so a failure here is only logged
	if user.Provider == clerkProvider {
		if _, err := clerkuser.Delete(c.Request.Context(), user.ProviderID); err != nil {
			log.Printf("users: delete clerk user %s: %v", user.ProviderID, err)
		}
	}
	c.Status(http.StatusNoContent)
}
