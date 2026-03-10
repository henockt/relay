package api

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/henockt/relay/internal/config"
	"github.com/henockt/relay/internal/email"
	"github.com/henockt/relay/internal/store"
	"gorm.io/gorm"
)

type Server struct {
	router     *gin.Engine
	cfg        *config.Config
	db         *gorm.DB
	userStore  *store.UserStore
	aliasStore *store.AliasStore
	sender     email.Sender
}

func NewServer(cfg *config.Config, db *gorm.DB, userStore *store.UserStore, aliasStore *store.AliasStore, sender email.Sender) *Server {
	s := &Server{
		router:     gin.Default(),
		cfg:        cfg,
		db:         db,
		userStore:  userStore,
		aliasStore: aliasStore,
		sender:     sender,
	}
	s.registerRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{s.cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	s.router.GET("/healthz", s.handleHealthz)
	s.router.GET("/readyz", s.handleReadyz)

	// proxy all non-API traffic to Next.js
	s.router.NoRoute(s.proxyToFrontend())

	api := s.router.Group("/api")
	{
		authGroup := api.Group("/auth")
		authGroup.GET("/google", s.handleGoogleLogin)
		authGroup.GET("/google/callback", s.handleGoogleCallback)
	}

	// inbound parse
	api.POST("/webhooks/email", s.handleInboundEmail)

	{
		protected := api.Group("/")
		protected.Use(s.authMiddleware())

		{
			users := protected.Group("/users/me")
			users.GET("", s.handleGetMe)
			users.DELETE("", s.handleDeleteMe)
		}

		{
			aliases := protected.Group("/aliases")
			aliases.GET("", s.handleListAliases)
			aliases.POST("", s.handleCreateAlias)
			aliases.PATCH("/:id", s.handleUpdateAlias)
			aliases.DELETE("/:id", s.handleDeleteAlias)
		}
	}
}

func (s *Server) handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleReadyz(c *gin.Context) {
	db, err := s.db.DB()
	if err != nil || db.Ping() != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// requests to the Next.js standalone
// server running on port 3000 inside the same container
func (s *Server) proxyToFrontend() gin.HandlerFunc {
	target, _ := url.Parse("http://127.0.0.1:3000")
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
