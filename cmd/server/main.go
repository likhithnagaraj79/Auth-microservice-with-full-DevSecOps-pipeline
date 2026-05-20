package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/likhithnagaraj79/auth-service/internal/auth"
	"github.com/likhithnagaraj79/auth-service/internal/database"
	"github.com/likhithnagaraj79/auth-service/internal/handlers"
	"github.com/likhithnagaraj79/auth-service/internal/middleware"
	"github.com/likhithnagaraj79/auth-service/internal/oauth"
	"github.com/likhithnagaraj79/auth-service/internal/rbac"
	"github.com/likhithnagaraj79/auth-service/pkg/config"
	"github.com/likhithnagaraj79/auth-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Server.Env)
	defer zap.L().Sync() //nolint:errcheck

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		zap.S().Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		zap.S().Fatalf("migrations failed: %v", err)
	}

	jwtSvc := auth.NewJWTService(&cfg.JWT)
	googleOAuth := oauth.NewGoogleProvider(&cfg.OAuth)

	authHandler := handlers.NewAuthHandler(db, jwtSvc, googleOAuth, cfg)
	userHandler := handlers.NewUserHandler(db)

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.AuditLogger(db))

	// Health
	r.GET("/health", userHandler.HealthCheck)

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.GET("/oauth/google", authHandler.GoogleLogin)
			authGroup.GET("/oauth/google/callback", authHandler.GoogleCallback)
		}

		protected := api.Group("/")
		protected.Use(middleware.AuthRequired(jwtSvc))
		{
			protected.GET("/me", authHandler.Me)
			protected.GET("/me/permissions", userHandler.Permissions)

			admin := protected.Group("/admin")
			admin.Use(middleware.RequirePermission(rbac.PermReadUsers))
			{
				admin.GET("/users", userHandler.ListUsers)
				admin.GET("/users/:id", userHandler.GetUser)
			}

			adminWrite := protected.Group("/admin")
			adminWrite.Use(middleware.RequirePermission(rbac.PermManageRoles))
			{
				adminWrite.PATCH("/users/:id/role", userHandler.UpdateRole)
			}

			adminDelete := protected.Group("/admin")
			adminDelete.Use(middleware.RequirePermission(rbac.PermDeleteUsers))
			{
				adminDelete.DELETE("/users/:id", userHandler.DeleteUser)
			}

			adminLogs := protected.Group("/admin")
			adminLogs.Use(middleware.RequirePermission(rbac.PermReadLogs))
			{
				adminLogs.GET("/audit-logs", userHandler.GetAuditLogs)
			}
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		zap.S().Infof("server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.S().Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.S().Errorf("forced shutdown: %v", err)
	}
	zap.S().Info("server stopped")
}
