package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rawdah/rawdah-api/internal/ai"
	"github.com/rawdah/rawdah-api/internal/config"
	"github.com/rawdah/rawdah-api/internal/handlers"
	"github.com/rawdah/rawdah-api/internal/mailer"
	"github.com/rawdah/rawdah-api/internal/middleware"
	"github.com/rawdah/rawdah-api/internal/migrate"
	"github.com/rawdah/rawdah-api/internal/push"
	"github.com/rawdah/rawdah-api/internal/repository"
	"github.com/rawdah/rawdah-api/internal/services"
	"github.com/rawdah/rawdah-api/internal/storage"
	"github.com/rawdah/rawdah-api/internal/ws"
)

func main() {
	// Setup zerolog
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	if cfg.CronSecret == "" {
		log.Warn().Msg("CRON_SECRET is not set — /cron/weekend-tasks endpoint is effectively disabled")
	}

	// Connect to DB
	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("connected to database")

	if cfg.AutoMigrate {
		migrationCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := migrate.Run(migrationCtx, db); err != nil {
			log.Fatal().Err(err).Msg("failed to run database migrations")
		}
		log.Info().Msg("database migrations applied")
	}

	// Create repositories
	authRepo := repository.NewAuthRepo(db)
	userRepo := repository.NewUserRepo(db)
	familyRepo := repository.NewFamilyRepo(db)
	taskRepo := repository.NewTaskRepo(db)
	rewardRepo := repository.NewRewardRepo(db)
	hadithRepo := repository.NewHadithRepo(db)
	prophetRepo := repository.NewProphetRepo(db)
	quranRepo := repository.NewQuranRepo(db)
	quizRepo := repository.NewQuizRepo(db)
	lessonRepo := repository.NewLessonRepo(db)
	msgRepo := repository.NewMessageRepo(db)
	rantRepo := repository.NewRantRepo(db)
	requestRepo := repository.NewRequestRepo(db)
	gameRepo := repository.NewGameRepo(db)
	dashRepo := repository.NewDashboardRepo(db)
	notifRepo := repository.NewNotificationRepo(db)
	pushRepo := repository.NewPushRepo(db)
	xpRepo := repository.NewXPRepo(db)
	recurringTaskRepo := repository.NewRecurringTaskRepo(db)

	// Setup WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Setup external services
	m := mailer.NewMailer(cfg)
	aiClient := ai.NewClient(cfg)

	var r2Client *storage.R2Client
	if cfg.R2AccountID != "" {
		r2Client, err = storage.NewR2Client(cfg)
		if err != nil {
			log.Warn().Err(err).Msg("failed to initialize R2 client, uploads will be disabled")
		}
	}

	pushSender := push.NewPushSender(cfg, pushRepo)

	// Create services
	xpSvc := services.NewXPService(xpRepo, quizRepo, lessonRepo)
	authSvc := services.NewAuthService(db, authRepo, userRepo, familyRepo, cfg)
	familySvc := services.NewFamilyService(familyRepo, xpRepo, m, cfg.AdultPlatformURL, cfg.KidsPlatformURL)
	taskSvc := services.NewTaskService(taskRepo, recurringTaskRepo, familyRepo, notifRepo, xpSvc, m, hub)
	rewardSvc := services.NewRewardService(rewardRepo)
	quizSvc := services.NewQuizService(quizRepo, hadithRepo, prophetRepo, quranRepo, familyRepo, notifRepo, xpSvc, aiClient, m, hub)
	assistantSvc := services.NewAssistantService(aiClient, familyRepo)
	lessonSvc := services.NewLessonService(lessonRepo, quizRepo, quranRepo, familyRepo, xpSvc, hub)
	msgSvc := services.NewMessageService(msgRepo, familyRepo, xpSvc, m, hub)
	rantSvc := services.NewRantService(rantRepo)
	reqSvc := services.NewRequestService(requestRepo, familyRepo, m, hub)
	gameSvc := services.NewGameService(gameRepo, userRepo, xpSvc, hub)
	dashSvc := services.NewDashboardService(dashRepo)
	notifSvc := services.NewNotificationService(notifRepo, pushSender, hub)

	// Create handlers
	authH := handlers.NewAuthHandler(authSvc, r2Client)
	familyH := handlers.NewFamilyHandler(familySvc, r2Client)
	taskH := handlers.NewTaskHandler(taskSvc)
	recurringH := handlers.NewRecurringTaskHandler(taskSvc, cfg)
	rewardH := handlers.NewRewardHandler(rewardSvc)
	hadithH := handlers.NewHadithHandler(hadithRepo)
	prophetH := handlers.NewProphetHandler(prophetRepo)
	quranH := handlers.NewQuranHandler(quranRepo)
	quizH := handlers.NewQuizHandler(quizSvc)
	assistantH := handlers.NewAssistantHandler(assistantSvc)
	lessonH := handlers.NewLessonHandler(lessonSvc)
	msgH := handlers.NewMessageHandler(msgSvc)
	rantH := handlers.NewRantHandler(rantSvc)
	reqH := handlers.NewRequestHandler(reqSvc)
	gameH := handlers.NewGameHandler(gameSvc)
	dashH := handlers.NewDashboardHandler(dashSvc)
	notifH := handlers.NewNotificationHandler(notifSvc)
	pushH := handlers.NewPushHandler(pushRepo)
	wsH := handlers.NewWSHandler(hub, cfg.AllowedOrigins)

	var uploadH *handlers.UploadHandler
	if r2Client != nil {
		uploadH = handlers.NewUploadHandler(r2Client, userRepo, familyRepo)
	}

	// Setup Gin router
	r := gin.New()
	r.Use(requestIDMiddleware())
	r.Use(zerologMiddleware())
	r.Use(recoveryMiddleware())
	r.Use(securityHeadersMiddleware())
	r.Use(corsMiddleware(cfg.AllowedOrigins))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})

	// Auth middleware instance
	authMW := middleware.AuthMiddleware(cfg)
	adultCanViewDashboard := middleware.AdultPermissionGuard(familyRepo, "view_dashboard")
	adultCanAssignTasks := middleware.AdultPermissionGuard(familyRepo, "assign_tasks")
	adultCanViewTasks := middleware.AdultPermissionGuard(familyRepo, "view_tasks")
	adultCanApproveRewards := middleware.AdultPermissionGuard(familyRepo, "approve_rewards")
	adultCanViewMessages := middleware.AdultPermissionGuard(familyRepo, "view_messages")
	adultCanManageQuizzes := middleware.AdultPermissionGuard(familyRepo, "manage_quizzes")
	adultCanManageLearn := middleware.AdultPermissionGuard(familyRepo, "manage_learn")
	adultCanRespondRequests := middleware.AdultPermissionGuard(familyRepo, "respond_requests")

	// WebSocket (requires auth via query param or header)
	r.GET("/ws", authMW, wsH.ServeWS)

	// Auth routes (public)
	auth := r.Group("/v1/auth")
	{
		auth.POST("/signup", authH.Signup)
		auth.POST("/signin", rateLimiter(10, time.Minute), authH.Signin)
		auth.POST("/child/signin", authH.ChildSignin)
		auth.POST("/refresh", authH.Refresh)
		auth.POST("/signout", authH.Signout)
	}

	// Protected routes
	v1 := r.Group("/v1", authMW)
	{
		// Auth (me)
		v1.GET("/auth/me", authH.Me)
		v1.PATCH("/auth/me/password", authH.ChangePassword)

		// Family
		v1.GET("/family", familyH.Get)
		v1.PATCH("/family", middleware.RoleGuard("parent"), familyH.Update)
		v1.GET("/family/members", familyH.ListMembers)
		v1.POST("/family/members", middleware.RoleGuard("parent"), familyH.CreateMember)
		v1.GET("/family/members/:id", familyH.GetMember)
		v1.PATCH("/family/members/:id", middleware.RoleGuard("parent"), familyH.UpdateMember)
		v1.DELETE("/family/members/:id", middleware.RoleGuard("parent"), familyH.DeactivateMember)
		v1.GET("/family/members/:id/rant-count", middleware.RoleGuard("parent"), familyH.RantCount)
		v1.GET("/family/access-control", middleware.RoleGuard("parent"), familyH.ListAccessControl)
		v1.PUT("/family/access-control/:grantee_id", middleware.RoleGuard("parent"), familyH.SetAccessControl)
		v1.DELETE("/family/access-control/:grantee_id", middleware.RoleGuard("parent"), familyH.RevokeAccessControl)

		// Tasks — recurring routes must be registered before /tasks/:id
		v1.GET("/tasks/recurring", middleware.RoleGuard("parent", "adult_relative"), adultCanAssignTasks, recurringH.List)
		v1.POST("/tasks/recurring", middleware.RoleGuard("parent", "adult_relative"), adultCanAssignTasks, recurringH.Create)
		v1.DELETE("/tasks/recurring/:id", middleware.RoleGuard("parent", "adult_relative"), adultCanAssignTasks, recurringH.Delete)

		// Tasks
		v1.GET("/tasks", adultCanViewTasks, taskH.List)
		v1.GET("/tasks/due-rewards", adultCanViewTasks, taskH.ListDueRewards)
		v1.POST("/tasks", middleware.RoleGuard("parent", "adult_relative"), adultCanAssignTasks, taskH.Create)
		v1.GET("/tasks/:id", adultCanViewTasks, taskH.Get)
		v1.PATCH("/tasks/:id", middleware.RoleGuard("parent", "adult_relative"), adultCanAssignTasks, taskH.Update)
		v1.DELETE("/tasks/:id", middleware.RoleGuard("parent", "adult_relative"), adultCanAssignTasks, taskH.Delete)
		v1.POST("/tasks/:id/start", middleware.RoleGuard("child"), taskH.Start)
		v1.POST("/tasks/:id/complete", middleware.RoleGuard("child"), taskH.Complete)
		v1.POST("/tasks/:id/request-reward", middleware.RoleGuard("child"), taskH.RequestReward)
		v1.POST("/tasks/:id/approve-reward", middleware.RoleGuard("parent", "adult_relative"), adultCanApproveRewards, taskH.ApproveReward)
		v1.POST("/tasks/:id/decline-reward", middleware.RoleGuard("parent", "adult_relative"), adultCanApproveRewards, taskH.DeclineReward)

		// Rewards
		v1.GET("/rewards", rewardH.List)
		v1.POST("/rewards", middleware.RoleGuard("parent", "adult_relative"), adultCanAssignTasks, rewardH.Create)
		v1.PATCH("/rewards/:id", middleware.RoleGuard("parent"), rewardH.Update)
		v1.DELETE("/rewards/:id", middleware.RoleGuard("parent"), rewardH.Delete)

		// Islamic content (public to all authenticated)
		v1.GET("/hadiths", hadithH.List)
		v1.GET("/hadiths/random", hadithH.Random)
		v1.GET("/hadiths/learned", middleware.RoleGuard("child"), hadithH.Learned)
		v1.GET("/hadiths/:id", hadithH.Get)
		v1.GET("/prophets", prophetH.List)
		v1.GET("/prophets/:id", prophetH.Get)
		v1.GET("/quran/verses", quranH.ListVerses)
		v1.GET("/quran/verses/:id", quranH.GetVerse)

		// Quizzes
		v1.POST("/quizzes/hadith/self", middleware.RoleGuard("child"), quizH.SelfAssignHadith)
		v1.POST("/quizzes/hadith", middleware.RoleGuard("parent", "adult_relative"), adultCanManageQuizzes, quizH.AssignHadith)
		v1.POST("/quizzes/prophet", middleware.RoleGuard("parent", "adult_relative"), adultCanManageQuizzes, quizH.AssignProphet)
		v1.POST("/quizzes/quran", middleware.RoleGuard("parent", "adult_relative"), adultCanManageQuizzes, quizH.AssignQuran)
		v1.POST("/quizzes/topic", middleware.RoleGuard("parent", "adult_relative"), adultCanManageQuizzes, quizH.AssignTopic)
		v1.GET("/quizzes", middleware.RoleGuard("parent", "adult_relative"), adultCanManageQuizzes, quizH.List)
		v1.GET("/quizzes/my", middleware.RoleGuard("child"), quizH.ListMine)
		v1.GET("/quizzes/:type/:id", adultCanManageQuizzes, quizH.Get)
		v1.POST("/quizzes/:type/:id/start", middleware.RoleGuard("child"), quizH.Start)
		v1.POST("/quizzes/:type/:id/submit", middleware.RoleGuard("child"), quizH.Submit)
		v1.POST("/ai/ask", rateLimiter(30, time.Minute), assistantH.Ask)

		// Quran lessons
		v1.GET("/lessons/quran", middleware.RoleGuard("parent", "adult_relative"), adultCanManageLearn, lessonH.ListLessons)
		v1.POST("/lessons/quran", middleware.RoleGuard("parent", "adult_relative"), adultCanManageLearn, lessonH.CreateLesson)
		v1.GET("/lessons/quran/my", middleware.RoleGuard("child"), lessonH.ListMyLessons)
		v1.GET("/lessons/quran/:id", middleware.RoleGuard("parent", "adult_relative"), adultCanManageLearn, lessonH.GetLesson)
		v1.POST("/lessons/quran/:id/complete", middleware.RoleGuard("child"), lessonH.CompleteLesson)

		// Learn content
		v1.GET("/learn", middleware.RoleGuard("parent", "adult_relative"), adultCanManageLearn, lessonH.ListLearnContent)
		v1.POST("/learn", middleware.RoleGuard("parent", "adult_relative"), adultCanManageLearn, lessonH.CreateLearnContent)
		v1.GET("/learn/my", middleware.RoleGuard("child"), lessonH.ListMyLearnContent)
		v1.POST("/learn/:id/complete", middleware.RoleGuard("child"), lessonH.CompleteLearnContent)

		// Messages
		v1.GET("/messages/conversations", adultCanViewMessages, msgH.Conversations)
		v1.GET("/messages/:user_id", adultCanViewMessages, msgH.Thread)
		v1.POST("/messages", adultCanViewMessages, msgH.Send)
		v1.PATCH("/messages/:id/read", adultCanViewMessages, msgH.MarkRead)

		// Rants (child only)
		v1.GET("/rants", middleware.RoleGuard("child"), rantH.List)
		v1.POST("/rants", middleware.RoleGuard("child"), rantH.Create)
		v1.GET("/rants/:id", middleware.RoleGuard("child"), rantH.Get)
		v1.PATCH("/rants/:id", middleware.RoleGuard("child"), rantH.Update)
		v1.DELETE("/rants/:id", middleware.RoleGuard("child"), rantH.Delete)

		// Requests
		v1.GET("/requests", reqH.List)
		v1.POST("/requests", middleware.RoleGuard("child"), reqH.Create)
		v1.GET("/requests/:id", reqH.Get)
		v1.POST("/requests/:id/respond", middleware.RoleGuard("parent", "adult_relative"), adultCanRespondRequests, reqH.Respond)

		// Games
		v1.GET("/games", gameH.ListAvailable)
		v1.POST("/games/sessions/start", middleware.RoleGuard("child"), gameH.StartSession)
		v1.POST("/games/sessions/:id/end", middleware.RoleGuard("child"), gameH.EndSession)
		v1.GET("/games/sessions", gameH.ListSessions)

		// Dashboard
		v1.GET("/dashboard/summary", middleware.RoleGuard("parent", "adult_relative"), adultCanViewDashboard, dashH.Summary)
		v1.GET("/dashboard/task-completion", middleware.RoleGuard("parent", "adult_relative"), adultCanViewDashboard, dashH.TaskCompletion)
		v1.GET("/dashboard/game-time", middleware.RoleGuard("parent", "adult_relative"), adultCanViewDashboard, dashH.GameTime)
		v1.GET("/dashboard/quiz-scores", middleware.RoleGuard("parent", "adult_relative"), adultCanViewDashboard, dashH.QuizScores)
		v1.GET("/dashboard/learn-progress", middleware.RoleGuard("parent", "adult_relative"), adultCanViewDashboard, dashH.LearnProgress)

		// Notifications
		v1.GET("/notifications", notifH.List)
		v1.PATCH("/notifications/read-all", notifH.ReadAll)
		v1.PATCH("/notifications/:id/read", notifH.ReadOne)

		// Push subscriptions
		v1.POST("/push/subscribe", pushH.Subscribe)
		v1.DELETE("/push/subscribe", pushH.Unsubscribe)

		// Upload
		if uploadH != nil {
			v1.POST("/upload/avatar", uploadH.PresignAvatar)
			v1.POST("/upload/avatar/confirm", uploadH.ConfirmAvatar)
			v1.POST("/upload/logo", middleware.RoleGuard("parent"), uploadH.PresignLogo)
			v1.POST("/upload/logo/confirm", middleware.RoleGuard("parent"), uploadH.ConfirmLogo)
		} else {
			v1.POST("/upload/avatar", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upload service not configured"})
			})
			v1.POST("/upload/avatar/confirm", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upload service not configured"})
			})
			v1.POST("/upload/logo", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upload service not configured"})
			})
			v1.POST("/upload/logo/confirm", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upload service not configured"})
			})
		}
	}

	// Cron trigger — no auth middleware
	cron := r.Group("/cron")
	cron.POST("/weekend-tasks", recurringH.TriggerWeekend)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info().Str("addr", addr).Msg("starting server")
	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}

func zerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		evt := log.Info()
		switch {
		case status >= http.StatusInternalServerError:
			evt = log.Error()
		case status >= http.StatusBadRequest:
			evt = log.Warn()
		}

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		evt = evt.
			Str("request_id", c.GetString("request_id")).
			Str("method", c.Request.Method).
			Str("route", route).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Str("status_text", http.StatusText(status)).
			Dur("latency", time.Since(start)).
			Int("bytes_out", c.Writer.Size()).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent())

		if userID := c.GetString("user_id"); userID != "" {
			evt = evt.Str("user_id", userID)
		}
		if familyID := c.GetString("family_id"); familyID != "" {
			evt = evt.Str("family_id", familyID)
		}
		if strings.TrimSpace(c.Request.URL.RawQuery) != "" {
			evt = evt.Bool("has_query", true)
		}
		if len(c.Errors) > 0 {
			errs := make([]string, 0, len(c.Errors))
			for _, e := range c.Errors {
				errs = append(errs, e.Error())
			}
			evt = evt.Strs("errors", errs)
		}

		evt.Msg("http_request")
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set("request_id", reqID)
		c.Header("X-Request-ID", reqID)
		c.Next()
	}
}

func recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		reqID := c.GetString("request_id")
		evt := log.Error().
			Str("request_id", reqID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Interface("panic", recovered).
			Bytes("stack", debug.Stack())
		if userID := c.GetString("user_id"); userID != "" {
			evt = evt.Str("user_id", userID)
		}
		if familyID := c.GetString("family_id"); familyID != "" {
			evt = evt.Str("family_id", familyID)
		}
		evt.Msg("panic_recovered")

		body := gin.H{
			"error": "The server encountered an unexpected error while processing your request.",
		}
		if reqID != "" {
			body["request_id"] = reqID
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, body)
	})
}

func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-XSS-Protection", "0")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowedSet := make(map[string]bool)
	for _, o := range allowedOrigins {
		allowedSet[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		originAllowed := origin != "" && allowedSet[origin]
		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Rant-Password")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "600")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			if origin != "" && !originAllowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Request origin is not allowed."})
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		if origin != "" && !originAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Request origin is not allowed."})
			return
		}
		c.Next()
	}
}

// Simple in-memory rate limiter
type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

var (
	rateLimitStore = make(map[string]*rateLimitEntry)
	rateLimitMu    sync.Mutex
)

func rateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rateLimitMu.Lock()
		for k, v := range rateLimitStore {
			if now.After(v.resetTime) {
				delete(rateLimitStore, k)
			}
		}
		entry, exists := rateLimitStore[ip]
		if !exists {
			rateLimitStore[ip] = &rateLimitEntry{
				count:     1,
				resetTime: now.Add(window),
			}
			rateLimitMu.Unlock()
			c.Next()
			return
		}

		entry.count++
		if entry.count > maxRequests {
			rateLimitMu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			return
		}
		rateLimitMu.Unlock()
		c.Next()
	}
}
