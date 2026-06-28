package api

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/soumabali/vexa/config"
	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/gateway"
	"github.com/soumabali/vexa/internal/hosts"
	"github.com/soumabali/vexa/internal/middleware"
	"github.com/soumabali/vexa/internal/models"
	"github.com/soumabali/vexa/internal/team"
	"github.com/soumabali/vexa/internal/terminal"
	"github.com/soumabali/vexa/internal/vault"
	"github.com/soumabali/vexa/internal/wireguard"

	"github.com/soumabali/vexa/internal/api/handlers"
)

func SetupRouter(cfg *config.Config, db *sql.DB, redisClient *redis.Client) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ErrorHandler())
	r.Use(middleware.RequestLogger())

	// Security middleware
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.AllowedOrigins))
	r.Use(middleware.RequestSizeLimit(cfg.MaxRequestSize))
	r.Use(middleware.NoCache())

	// Rate limiting
	r.Use(middleware.APIRateLimiter(redisClient))

	// Services
	jwtManager := auth.NewJWTManager([]byte(cfg.JWTSecret), []byte(cfg.JWTRefreshSecret), cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	mfaService := auth.NewMFAService("vexa", cfg.EncryptionKey)
	sessionStore := auth.NewSessionStore(redisClient, 30*time.Minute, 8*time.Hour)

	// User service for authentication
	userService := auth.NewUserService(db, redisClient, cfg.EncryptionKey)

	// Audit logger
	auditLogger, err := audit.NewLogger(db, "logs/audit.log", []byte(cfg.EncryptionKey))
	if err != nil {
		panic(err)
	}

	// Credential service (uses shared vault instance)
	credService := vault.NewCredentialService(db, redisClient, auditLogger)

	// Repositories
	hostRepo := hosts.NewRepository(db)

	// Gateway
	sshGateway := gateway.NewSSHGateway(&gateway.SSHGatewayConfig{
		MaxConnectionsPerHost: 10,
		ConnectionTimeout:     10 * time.Second,
		IdleTimeout:           5 * time.Minute,
		KeepAliveInterval:     30 * time.Second,
	})

	rdpGateway := gateway.NewRDPGateway(&gateway.RDPGatewayConfig{
		ConnectionTimeout: 10 * time.Second,
		MaxSessions:       100,
	})

	vncGateway := gateway.NewVNCGateway(&gateway.VNCGatewayConfig{
		ConnectionTimeout: 10 * time.Second,
		MaxSessions:       100,
		PasswordLength:    8,
	})

	_ = sshGateway
	_ = rdpGateway
	_ = vncGateway

	// Session manager
	terminalManager := terminal.NewSessionManager()
	_ = terminalManager

	// Login rate limiter with progressive backoff
	loginRateLimiter := middleware.NewLoginRateLimiter(redisClient)

	// WebAuthn service
	webauthnService, err := auth.NewWebAuthnService(db, cfg.WebAuthnRPID, cfg.WebAuthnRPOrigin, cfg.WebAuthnRPDisplayName)
	if err != nil {
		panic(err)
	}
	webauthnHandler := handlers.NewWebAuthnHandler(webauthnService, userService, jwtManager, sessionStore, auditLogger)

	// Handlers
	authHandler := handlers.NewAuthHandler(userService, jwtManager, mfaService, sessionStore, auditLogger, loginRateLimiter)
	hostHandler := handlers.NewHostHandler(hostRepo, auditLogger)
	credHandler := handlers.NewCredentialHandler(credService, auditLogger)
	shareRepo := vault.NewShareRepository(db)
	shareHandler := vault.NewShareHandler(shareRepo)
	teamService := team.NewTeamService(db, auditLogger)
	teamHandler := handlers.NewTeamHandler(teamService, auditLogger)
	terminalHandler := handlers.NewTerminalHandler(terminalManager, jwtManager, sessionStore, cfg.AllowedOrigins, hostRepo, credService, sshGateway, auditLogger)
	auditHandler := handlers.NewAuditHandler(auditLogger)
	adminWireguardHandler := handlers.NewAdminWireguardHandler(auditLogger, "")
	gatewayHandler := handlers.NewGatewayHandler(sshGateway, rdpGateway, vncGateway, auditLogger)
	userHandler := handlers.NewUserHandler(userService, auditLogger)

	// Public routes (no auth required)
	public := r.Group("/api/v1")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", loginRateLimiter.Middleware(), authHandler.Login)
		public.POST("/auth/mfa/verify", authHandler.VerifyMFA)
		public.POST("/auth/refresh", authHandler.RefreshToken)

		// WebAuthn login (public, passwordless)
		public.POST("/auth/webauthn/login/begin", webauthnHandler.LoginBegin)
		public.POST("/auth/webauthn/login/finish", webauthnHandler.LoginFinish)
	}

	// Authenticated routes
	authenticated := r.Group("/api/v1")
	authenticated.Use(middleware.JWTAuth(jwtManager, sessionStore))
	{
		// Auth
		authenticated.POST("/auth/logout", authHandler.Logout)
		authenticated.POST("/auth/mfa/setup", authHandler.SetupMFA)
		authenticated.POST("/auth/mfa/enable", authHandler.VerifyMFAEnable)
		authenticated.POST("/auth/mfa/backup-codes/regenerate", authHandler.RegenerateBackupCodes)
		authenticated.DELETE("/auth/mfa/disable", authHandler.DisableMFA)
		authenticated.GET("/auth/sessions", authHandler.GetActiveSessions)
		authenticated.POST("/auth/sessions/revoke", authHandler.RevokeSession)

		// WebAuthn registration and management
		authenticated.POST("/auth/webauthn/register/begin", webauthnHandler.RegisterBegin)
		authenticated.POST("/auth/webauthn/register/finish", webauthnHandler.RegisterFinish)
		authenticated.GET("/auth/webauthn/credentials", webauthnHandler.ListCredentials)
		authenticated.DELETE("/auth/webauthn/credentials/:id", webauthnHandler.DeleteCredential)
		authenticated.PATCH("/auth/webauthn/credentials/:id", webauthnHandler.UpdateCredential)

		// Vault
		authenticated.POST("/vault/unlock", credHandler.Unlock)
		authenticated.POST("/vault/lock", credHandler.Lock)
		authenticated.GET("/vault/status", credHandler.Status)
		authenticated.POST("/vault/key/rotate", credHandler.RotateKey)

		// Credentials
		authenticated.POST("/vault/credentials", credHandler.Create)
		authenticated.GET("/vault/credentials", credHandler.List)
		authenticated.GET("/vault/credentials/:id", credHandler.Get)
		authenticated.PATCH("/vault/credentials/:id", credHandler.Update)
		authenticated.DELETE("/vault/credentials/:id", credHandler.Delete)
		authenticated.GET("/vault/credentials/:id/decrypt", credHandler.Decrypt)
		authenticated.POST("/vault/credentials/:id/share", credHandler.Share)
		authenticated.DELETE("/vault/credentials/:id/share", credHandler.Unshare)

		// Vault credential sharing (E2E encrypted, ShareHandler)
		vault.RegisterShareRoutes(authenticated, shareHandler)

		// Teams
		authenticated.POST("/teams", teamHandler.Create)
		authenticated.GET("/teams", teamHandler.List)
		authenticated.GET("/teams/:id", teamHandler.Get)
		authenticated.PATCH("/teams/:id", teamHandler.Update)
		authenticated.DELETE("/teams/:id", teamHandler.Delete)
		authenticated.GET("/teams/:id/members", teamHandler.ListMembers)
		authenticated.POST("/teams/:id/members", teamHandler.AddMember)
		authenticated.PATCH("/teams/:id/members/:user_id", teamHandler.UpdateMemberRole)
		authenticated.DELETE("/teams/:id/members/:user_id", teamHandler.RemoveMember)

		// Hosts
		authenticated.POST("/hosts", hostHandler.Create)
		authenticated.GET("/hosts", hostHandler.List)
		authenticated.GET("/hosts/:id", hostHandler.Get)
		authenticated.PATCH("/hosts/:id", hostHandler.Update)
		authenticated.DELETE("/hosts/:id", hostHandler.Delete)
		authenticated.GET("/hosts/:id/health", hostHandler.HealthCheck)

		// Terminal
		authenticated.GET("/ws/terminal", terminalHandler.HandleTerminal)
		authenticated.GET("/sessions", terminalHandler.ListSessions)
		authenticated.DELETE("/sessions/:id", terminalHandler.CloseSession)

		// Users (self)
		authenticated.GET("/users/me", userHandler.GetProfile)
		authenticated.PATCH("/users/me", userHandler.UpdateProfile)
		authenticated.POST("/users/me/password", userHandler.ChangePassword)

		// Audit (admin only)
		admin := authenticated.Group("/admin")
		admin.Use(middleware.RequireRole(string(models.RoleAdmin)))
		{
			admin.GET("/audit", auditHandler.Query)
			admin.GET("/audit/export", auditHandler.Export)
			admin.POST("/audit/log", auditHandler.LogEvent)
			admin.GET("/users", userHandler.ListUsers)
			admin.POST("/users", userHandler.CreateUser)
			admin.GET("/users/:id", userHandler.GetUser)
			admin.PATCH("/users/:id", userHandler.UpdateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)
			// WireGuard rotation (admin only) — P4 #3
			admin.POST("/wg/rotate", adminWireguardHandler.RotateWireguardKeys)
		}
	}

	// WireGuard service
	wgCfg := wireguard.ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		RotationDays:      90,
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "wg",
	}
	wgService := wireguard.NewWireGuardService(db, auditLogger, wgCfg)
	tunnelHandler := handlers.NewTunnelHandler(wgService)
	tunnelWSHandler := handlers.NewTunnelWSHandler(wgService)

	// WireGuard tunnel routes
	authenticated.POST("/tunnels", tunnelHandler.Create)
	authenticated.GET("/tunnels", tunnelHandler.List)
	authenticated.GET("/tunnels/:id", tunnelHandler.Get)
	authenticated.PATCH("/tunnels/:id", tunnelHandler.Update)
	authenticated.DELETE("/tunnels/:id", tunnelHandler.Delete)
	authenticated.POST("/tunnels/:id/rotate", tunnelHandler.Rotate)
	authenticated.GET("/tunnels/:id/config", tunnelHandler.Config)
	authenticated.POST("/tunnels/:id/enable", tunnelHandler.Enable)
	authenticated.POST("/tunnels/:id/disable", tunnelHandler.Disable)
	authenticated.GET("/tunnels/:id/stats", tunnelHandler.Stats)
	authenticated.GET("/ws/tunnels", tunnelWSHandler.HandleTunnelWS)

	// Gateway proxy routes (for RDP/VNC/Web protocols)
	gatewayRoutes := r.Group("/api/v1/gateway")
	gatewayRoutes.Use(middleware.JWTAuth(jwtManager, sessionStore))
	{
		gatewayRoutes.POST("/ssh/connect", gatewayHandler.SSHConnect)
		gatewayRoutes.POST("/rdp/connect", gatewayHandler.RDPConnect)
		gatewayRoutes.POST("/vnc/connect", gatewayHandler.VNCConnect)
	}

	// Health check - Kubernetes liveness probe
	healthHandler := handlers.NewHealthHandler(db, redisClient, os.Getenv("DATA_DIR"))
	metricsHandler := handlers.NewMetricsHandler()
	r.Use(metricsHandler.PrometheusMiddleware())
	r.GET("/health/live", healthHandler.Live)
	r.GET("/health/ready", healthHandler.Ready)

	// Legacy /health redirect
	r.GET("/health", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/health/ready")
	})

	return r
}
