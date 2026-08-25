// Reescrito desde ../usbi/backend/main.go: USBI-Anon abre DOS conexiones
// PostgreSQL (identidad y principal) en vez de una — ver
// plan/00_Plan_maestro.md §2 y plan/02_Backend.md §3. Cada servicio recibe
// solo las Queries de la(s) base(s) que realmente necesita: auth y
// maintenance reciben ambas (ident + main); sync/levels/devices/incidents/
// dbmaint siguen recibiendo únicamente main, sin cambio frente a ../usbi.
package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/altair/usbi-anon-backend/internal/auth"
	"github.com/altair/usbi-anon-backend/internal/config"
	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/dbmaint"
	"github.com/altair/usbi-anon-backend/internal/devices"
	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/incidents"
	"github.com/altair/usbi-anon-backend/internal/levels"
	"github.com/altair/usbi-anon-backend/internal/mailer"
	"github.com/altair/usbi-anon-backend/internal/maintenance"
	"github.com/altair/usbi-anon-backend/internal/repository"
	syncSvc "github.com/altair/usbi-anon-backend/internal/sync"
	"github.com/altair/usbi-anon-backend/internal/transport"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {
	config.LoadEnvironment()

	// ── Structured logging (B4) ───────────────────────────────────────────────
	logger := newLogger()
	slog.SetDefault(logger)
	// The background schedulers take a *log.Logger; bridge it through slog so
	// their output is structured too.
	schedulerLogger := slog.NewLogLogger(logger.Handler(), slog.LevelInfo)

	// Root context cancelled on SIGINT/SIGTERM so the server and the background
	// schedulers all drain cleanly on `systemctl restart` (B4).
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Required environment variables ────────────────────────────────────────
	mainDBURL := config.DatabaseURL()
	identDBURL := config.IdentityDatabaseURL()
	jwtSecret := config.RequireSecret("JWT_SECRET")
	encryptionKey := config.RequireSecret("PGP_ENCRYPTION_KEY")
	blindIndexSecret := config.RequireSecret("BLIND_INDEX_SECRET")
	hmacSecret := config.RequireSecret("HMAC_SECRET")

	// Optional with defaults
	port := config.GetEnv("SERVER_PORT", "8088")
	allowedOrigin := config.GetEnv("CORS_ALLOWED_ORIGIN", "")

	// ── Tutor double opt-in (A1): mailer + verification link ──────────────────
	appMailer := buildMailer()
	tutorVerifyURL := ""
	if base := config.GetEnv("PUBLIC_API_BASE_URL", ""); base != "" {
		tutorVerifyURL = strings.TrimRight(base, "/") + "/api/v1/auth/tutor-consent/verify"
	}
	tutorTokenTTL := config.GetDurationEnv("TUTOR_CONSENT_TOKEN_TTL", 24*time.Hour)

	accessExpiryStr := config.GetEnv("JWT_ACCESS_EXPIRY_MINUTES", "15")
	accessExpiryMinutes, err := strconv.Atoi(accessExpiryStr)
	if err != nil {
		log.Fatalf("[FATAL] Invalid JWT_ACCESS_EXPIRY_MINUTES: %v", err)
	}

	// ── Database connections — DOS pools independientes ───────────────────────
	mainDB := openPool(mainDBURL, "principal", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME", 10, 2)
	defer mainDB.Close()

	identDB := openPool(identDBURL, "identidad", "IDENT_DB_MAX_OPEN_CONNS", "IDENT_DB_MAX_IDLE_CONNS",
		"IDENT_DB_CONN_MAX_LIFETIME", "IDENT_DB_CONN_MAX_IDLE_TIME", 5, 1)
	defer identDB.Close()

	// ── Repository & Services ─────────────────────────────────────────────────
	mainQueries := repository.New(mainDB)
	identQueries := identityrepo.New(identDB)

	tokenCfg := crypto.TokenConfig{
		Secret:       []byte(jwtSecret),
		AccessExpiry: time.Duration(accessExpiryMinutes) * time.Minute,
	}

	authSvc := auth.NewService(identQueries, mainQueries, auth.Config{
		EncryptionKey:               encryptionKey,
		BlindIndexSecret:            []byte(blindIndexSecret),
		HMACSecret:                  []byte(hmacSecret),
		TokenConfig:                 tokenCfg,
		MaxConcurrentPasswordHashes: int(config.GetInt32Env("MAX_CONCURRENT_PASSWORD_HASHES", 2)),
		Mailer:                      appMailer,
		TutorConsentVerifyURL:       tutorVerifyURL,
		TutorConsentTokenTTL:        tutorTokenTTL,
	})

	syncService := syncSvc.NewService(mainQueries, []byte(hmacSecret))
	levelsSvc := levels.NewService(mainQueries)
	devicesSvc := devices.NewService(mainQueries)
	incidentsSvc := incidents.NewService(mainQueries, []byte(hmacSecret))
	if config.GetBoolEnv("LEGAL_MAINTENANCE_ENABLED", false) {
		maintenanceSvc := maintenance.NewService(identQueries, mainQueries, maintenance.Config{
			EncryptionKey:        encryptionKey,
			BlindIndexSecret:     []byte(blindIndexSecret),
			PendingTutorTTL:      config.GetDurationEnv("PENDING_TUTOR_TTL", 48*time.Hour),
			InactiveSuspendAfter: config.GetDurationEnv("INACTIVE_SUSPEND_AFTER", 365*24*time.Hour),
			SuspendedCancelAfter: config.GetDurationEnv("SUSPENDED_CANCEL_AFTER", 30*24*time.Hour),
			ArcoStuckAfter:       config.GetDurationEnv("ARCO_STUCK_AFTER", 15*time.Minute),
			BatchSize:            config.GetInt32Env("LEGAL_MAINTENANCE_BATCH_SIZE", 100),
		})
		maintenance.StartScheduler(
			rootCtx,
			maintenanceSvc,
			config.GetDurationEnv("LEGAL_MAINTENANCE_INTERVAL", 24*time.Hour),
			schedulerLogger,
		)
		slog.Info("legal maintenance scheduler enabled")
	}

	// Structural DB concern, independent of legal/privacy retention above —
	// keeps level_attempts/daily_streak (base principal) supplied with future
	// yearly partitions so inserts never hit the DEFAULT partition in
	// practice. No aplica a la base de identidad: ninguna de sus tablas está
	// particionada.
	if config.GetBoolEnv("DB_PARTITION_MAINTENANCE_ENABLED", true) {
		dbmaintSvc := dbmaint.NewService(mainDB)
		dbmaint.StartScheduler(
			rootCtx,
			dbmaintSvc,
			config.GetDurationEnv("DB_PARTITION_MAINTENANCE_INTERVAL", 24*time.Hour),
			schedulerLogger,
		)
		slog.Info("db partition maintenance scheduler enabled")
	}

	// ── Router wiring ─────────────────────────────────────────────────────────
	r := chi.NewRouter()
	stopRateLimiters := transport.SetupRoutes(r, transport.RouterDependencies{
		AuthHandler:      auth.NewHandler(authSvc),
		SyncHandler:      syncSvc.NewHandler(syncService),
		LevelsHandler:    levels.NewHandler(levelsSvc),
		DevicesHandler:   devices.NewHandler(devicesSvc),
		IncidentsHandler: incidents.NewHandler(incidentsSvc),
		ReadyCheck:       readyCheck(mainDB, identDB),
		TokenCfg:         tokenCfg,
		IdentQueries:     identQueries,
		AllowedOrigin:    allowedOrigin,
		MaxBodyBytes:     int64(config.GetInt32Env("API_MAX_BODY_BYTES", 6*1024*1024)),
		// Only trust proxy-forwarded IP headers once a reverse proxy in front
		// of this service is confirmed to strip/set them itself (see DEPLOYMENT.md).
		TrustProxyHeaders: config.GetBoolEnv("TRUST_PROXY_HEADERS", false),
		// Per-request deadline so a slow query or a blocked advisory lock can't
		// pin a goroutine and one of the few pool connections forever (B5).
		RequestTimeout: config.GetDurationEnv("REQUEST_TIMEOUT", 20*time.Second),
	})
	defer stopRateLimiters()

	// ── TLS 1.2+ (RF69 — TLS 1.0/1.1 explicitly disabled) ────────────────────
	// Only applies when TLS_CERT_FILE/TLS_KEY_FILE are set below and this
	// process terminates TLS itself. When they're unset (the default, and the
	// documented production setup in DEPLOYMENT.md), a reverse proxy (Nginx)
	// is expected to terminate TLS and this server speaks plain HTTP to it.
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		PreferServerCipherSuites: true,
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		TLSConfig:    tlsCfg,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	certFile := config.GetEnv("TLS_CERT_FILE", "")
	keyFile := config.GetEnv("TLS_KEY_FILE", "")

	// Serve in the background so main can wait for a shutdown signal (B4).
	serverErr := make(chan error, 1)
	go func() {
		if certFile != "" && keyFile != "" {
			slog.Info("server listening", "addr", ":"+port, "tls", true)
			serverErr <- server.ListenAndServeTLS(certFile, keyFile)
		} else {
			slog.Info("server listening", "addr", ":"+port, "tls", false,
				"note", "expects TLS termination by a reverse proxy")
			serverErr <- server.ListenAndServe()
		}
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Server startup failed: %v", err)
		}
	case <-rootCtx.Done():
		// Stop intercepting signals so a second Ctrl-C force-kills if draining hangs.
		stop()
		slog.Info("shutdown signal received; draining connections")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
		} else {
			slog.Info("server stopped cleanly")
		}
	}
}

// openPool abre y valida (Ping) un pool de conexiones PostgreSQL, aplicando
// límites configurables por variables de entorno con nombres distintos para
// cada base (label solo identifica los mensajes de log). defaultMaxOpen para
// la base de identidad es deliberadamente más bajo que el de la principal:
// recibe muchas menos consultas por request (solo auth toca esta base en el
// camino síncrono de cada login/refresh) — ver plan/02_Backend.md §3.5.
func openPool(dsn, label, maxOpenVar, maxIdleVar, lifetimeVar, idleTimeVar string, defaultMaxOpen, defaultMaxIdle int32) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("[FATAL] Opening %s database: %v", label, err)
	}

	maxOpen := config.GetInt32Env(maxOpenVar, defaultMaxOpen)
	maxIdle := config.GetInt32Env(maxIdleVar, defaultMaxIdle)
	config.CheckConnPoolBounds(maxIdle, maxOpen)
	db.SetMaxOpenConns(int(maxOpen))
	db.SetMaxIdleConns(int(maxIdle))
	db.SetConnMaxLifetime(config.GetDurationEnv(lifetimeVar, 30*time.Minute))
	db.SetConnMaxIdleTime(config.GetDurationEnv(idleTimeVar, 5*time.Minute))

	if err := db.Ping(); err != nil {
		log.Fatalf("[FATAL] %s database unreachable: %v", label, err)
	}
	log.Printf("[INFO] %s database connection established", label)
	return db
}

// readyCheck pings AMBAS bases: /health/ready solo debe reportar "ready"
// cuando el servicio puede de verdad autenticar (identidad) y servir
// progreso (principal), no solo una de las dos — y el error indica cuál de
// las dos falló (criterio 5 de plan/02_Backend.md §8), sin filtrar el error
// crudo del driver (podría incluir fragmentos del DSN).
func readyCheck(mainDB, identDB *sql.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := mainDB.PingContext(ctx); err != nil {
			return errors.New("base principal inalcanzable")
		}
		if err := identDB.PingContext(ctx); err != nil {
			return errors.New("base de identidad inalcanzable")
		}
		return nil
	}
}

// newLogger builds the structured slog logger from LOG_LEVEL (debug|info|warn|
// error, default info) and LOG_FORMAT (json|text, default json).
func newLogger() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(config.GetEnv("LOG_LEVEL", "info")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level}
	if strings.ToLower(config.GetEnv("LOG_FORMAT", "json")) == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

// buildMailer constructs the transactional mailer from the environment. When
// SMTP_HOST is unset it falls back to a dev LogMailer (which only logs the
// verification link) and warns loudly — that fallback must never be used in
// production, where tutor consent links must actually be delivered by email.
func buildMailer() mailer.Mailer {
	host := config.GetEnv("SMTP_HOST", "")
	if host == "" {
		log.Println("[WARN] SMTP_HOST not set — using dev LogMailer; tutor verification " +
			"links will only appear in the logs. Configure SMTP before production.")
		return mailer.NewLogMailer(log.Default())
	}
	return mailer.NewSMTPMailer(mailer.Config{
		Host:        host,
		Port:        config.GetEnv("SMTP_PORT", "587"),
		Username:    config.GetEnv("SMTP_USER", ""),
		Password:    config.GetEnv("SMTP_PASSWORD", ""),
		From:        config.GetEnv("SMTP_FROM", ""),
		ImplicitTLS: config.GetBoolEnv("SMTP_IMPLICIT_TLS", false),
	})
}
