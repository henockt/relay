package api

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"github.com/henockt/relay/internal/auth"
	"github.com/henockt/relay/internal/models"
	"gorm.io/gorm"
)

const clerkProvider = "clerk"

var errEmailNotAllowed = errors.New("email domain not allowed")

func (s *Server) authMiddleware() gin.HandlerFunc {
	// verifies the Clerk session token and caches the signing key.
	// failures fall through with no claims so we answer in our own format
	verify := clerkhttp.WithHeaderAuthorization(
		clerkhttp.AuthorizationFailureHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})),
	)

	return func(c *gin.Context) {
		var claims *clerk.SessionClaims
		verify(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			claims, _ = clerk.SessionClaimsFromContext(r.Context())
		})).ServeHTTP(c.Writer, c.Request)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user, err := s.resolveUser(c.Request.Context(), claims.Subject)
		if err != nil {
			if errors.Is(err, errEmailNotAllowed) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "email_not_allowed"})
				return
			}
			log.Printf("auth: resolve user %s: %v", claims.Subject, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not verify user"})
			return
		}

		c.Set("userID", user.ID)
		c.Next()
	}
}

// resolveUser maps a Clerk user to a relay user. Only the first request
// from a new Clerk user hits the Clerk API, after that it's one lookup.
func (s *Server) resolveUser(ctx context.Context, clerkID string) (*models.User, error) {
	user, err := s.userStore.FindByProvider(clerkProvider, clerkID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	cu, err := clerkuser.Get(ctx, clerkID)
	if err != nil {
		return nil, err
	}
	email := primaryEmail(cu)
	if email == "" || !auth.AllowedEmail(email) {
		return nil, errEmailNotAllowed
	}

	// accounts created before Clerk (Google sign-in) are linked by email
	user, err = s.userStore.FindByEmail(email)
	switch {
	case err == nil:
		if err := s.userStore.UpdateProvider(user.ID, clerkProvider, clerkID); err != nil {
			return nil, err
		}
		return user, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	user = &models.User{Email: email, Provider: clerkProvider, ProviderID: clerkID}
	if err := s.userStore.Create(user); err != nil {
		// a concurrent first request may have created it already
		if existing, findErr := s.userStore.FindByProvider(clerkProvider, clerkID); findErr == nil {
			return existing, nil
		}
		return nil, err
	}
	return user, nil
}

// primaryEmail returns the user's verified primary address, or "".
func primaryEmail(u *clerk.User) string {
	if u.PrimaryEmailAddressID == nil {
		return ""
	}
	for _, e := range u.EmailAddresses {
		if e.ID == *u.PrimaryEmailAddressID {
			if e.Verification == nil || e.Verification.Status != "verified" {
				return ""
			}
			return e.EmailAddress
		}
	}
	return ""
}
