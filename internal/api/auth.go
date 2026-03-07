package api

// Standard OAuth login flow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/henockt/relay/internal/auth"
	"github.com/henockt/relay/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func (s *Server) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  s.cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
}

func (s *Server) handleGoogleLogin(c *gin.Context) {
	state := randomState()
	c.SetCookie("oauth_state", state, 300, "/", "", s.cfg.SecureCookies, true)
	c.Redirect(http.StatusTemporaryRedirect, s.googleOAuthConfig().AuthCodeURL(state))
}

func (s *Server) handleGoogleCallback(c *gin.Context) {
	// anti-CSRF
	if err := verifyState(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state mismatch"})
		return
	}

	// exchange 'code' for access token
	cfg := s.googleOAuthConfig()
	oauthToken, err := cfg.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code exchange failed"})
		return
	}

	// fetch user info from Google API
	resp, err := cfg.Client(context.Background(), oauthToken).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user info"})
		return
	}
	defer resp.Body.Close()

	var info struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
		return
	}

	jwtToken, err := s.upsertUserAndIssue("google", info.ID, info.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth failed"})
		return
	}

	// redirect with a query param
	// change to cookie later
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/callback?token=%s", s.cfg.FrontendURL, jwtToken))
}

// finds or creates the user and returns a signed JWT.
func (s *Server) upsertUserAndIssue(provider, providerID, email string) (string, error) {
	user, err := s.userStore.FindByProvider(provider, providerID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return "", err
		}
		// new user -> create the user
		user = &models.User{
			Email:      email,
			Provider:   provider,
			ProviderID: providerID,
		}
		if err := s.userStore.Create(user); err != nil {
			return "", err
		}
	}
	return auth.Issue(user.ID, s.cfg.JWTSecret)
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}

func verifyState(c *gin.Context) error {
	cookie, err := c.Cookie("oauth_state")
	if err != nil || cookie != c.Query("state") {
		return fmt.Errorf("state mismatch")
	}
	return nil
}
