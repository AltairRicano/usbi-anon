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
	"github.com/altair/usbi-anon-backend/internal/incidents"
	"github.com/altair/usbi-anon-backend/internal/interestlinks"
	"github.com/altair/usbi-anon-backend/internal/levels"
	"github.com/altair/usbi-anon-backend/internal/maintenance"
	"github.com/altair/usbi-anon-backend/internal/quiz"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/altair/usbi-anon-backend/internal/suggestions"
	syncSvc "github.com/altair/usbi-anon-backend/internal/sync"
	"github.com/altair/usbi-anon-backend/internal/transport"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {
	config.LoadEnvironment()

	logger := newLogger()
	slog.SetDefault(logger)

	schedulerLogger := slog.NewLogLogger(logger.Handler(), slog.LevelInfo)

	// Root context cancelled on SIGINT/SIGTERM so the server and the background
	// schedulers all drain cleanly on `systemctl restart`.
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Required environment variables ────────────────
	dbURL := config.DatabaseURL()
	moderatorDBURL := config.ModeratorDatabaseURL()
	dbmaintDBURL := config.DBMaintDatabaseURL()
	jwtSecret := config.RequireSecret("JWT_SECRET")
	// HMACSecret firma el token de registro (registration_token.go), el sello
	// de aceptación del aviso de privacidad y los tokens de refresh — el único
	// secreto de firmado que sobrevive sin cifrado ni blind index.
	hmacSecret := config.RequireSecret("HMAC_SECRET")

	// Optional with defaults
	port := config.GetEnv("SERVER_PORT", "8088")
	allowedOrigin := config.GetEnv("CORS_ALLOWED_ORIGIN", "")

	accessExpiryStr := config.GetEnv("JWT_ACCESS_EXPIRY_MINUTES", "15")
	accessExpiryMinutes, err := strconv.Atoi(accessExpiryStr)
	if err != nil {
		log.Fatalf("[FATAL] Invalid JWT_ACCESS_EXPIRY_MINUTES: %v", err)
	}

	// ── Database connections — TRES pools (F3, 2026-09-09) ────────────────────
	// player (usbi_app): autoservicio de jugador — login, registro, progreso,
	// dispositivos, cancelación de cuenta. moderator (usbi_moderador): todo lo
	// que ya vivía detrás de un guard de rol admin en Go (gestión de
	// contenido, banco de preguntas, incidentes de seguridad). dbmaint
	// (usbi_dbmaint): sin ningún GRANT de tabla, solo EXECUTE sobre la función
	// que mantiene las particiones anuales — ver internal/dbmaint.
	playerDB := openPool(dbURL, "jugador (usbi_app)", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME", 10, 2)
	defer playerDB.Close()

	moderatorDB := openPool(moderatorDBURL, "moderador (usbi_moderador)", "DB_MODERATOR_MAX_OPEN_CONNS", "DB_MODERATOR_MAX_IDLE_CONNS",
		"DB_MODERATOR_CONN_MAX_LIFETIME", "DB_MODERATOR_CONN_MAX_IDLE_TIME", 5, 1)
	defer moderatorDB.Close()

	dbmaintDB := openPool(dbmaintDBURL, "mantenimiento de particiones (usbi_dbmaint)", "DB_DBMAINT_MAX_OPEN_CONNS", "DB_DBMAINT_MAX_IDLE_CONNS",
		"DB_DBMAINT_CONN_MAX_LIFETIME", "DB_DBMAINT_CONN_MAX_IDLE_TIME", 2, 1)
	defer dbmaintDB.Close()

	// ── Repository & Services ─────────────────────────────────────────────────
	playerQueries := repository.New(playerDB)
	moderatorQueries := repository.New(moderatorDB)

	tokenCfg := crypto.TokenConfig{
		Secret:       []byte(jwtSecret),
		AccessExpiry: time.Duration(accessExpiryMinutes) * time.Minute,
	}

	// internal/auth mezcla autoservicio de jugador y administración de
	// cuentas en un solo Service a propósito — accounts no tiene Row-Level
	// Security, así que separar esas rutas por pool no añadía ninguna
	// restricción real (decisión registrada en estado_proyecto.md
	// 2026-09-02). Corre entero sobre el pool de jugador.
	quizPlayerSvc := quiz.NewPlayerService(playerQueries)
	quizAdminSvc := quiz.NewAdminService(moderatorQueries)
	authSvc := auth.NewService(playerQueries, quizPlayerSvc, auth.Config{
		HMACSecret:                  []byte(hmacSecret),
		TokenConfig:                 tokenCfg,
		MaxConcurrentPasswordHashes: int(config.GetInt32Env("MAX_CONCURRENT_PASSWORD_HASHES", 2)),
		StaffPrivacyNoticeVersion:   config.GetEnv("STAFF_PRIVACY_NOTICE_VERSION", ""),
	})

	// sync y devices son autoservicio de jugador puro (sincronización offline
	// de progreso propio) — nunca los dispara un admin, así que van enteros
	// sobre el pool de jugador.
	syncService := syncSvc.NewService(playerQueries, []byte(hmacSecret))
	levelsPlayerSvc := levels.NewPlayerService(playerQueries)
	levelsAdminSvc := levels.NewAdminService(moderatorQueries)
	devicesSvc := devices.NewService(playerQueries)
	// security_incidents no tiene ningún GRANT para usbi_app (REVOKE ALL,
	// 00_roles_unificado.sql) — el endpoint es admin-only pese al comentario
	// "/admin/" en la ruta, así que corre sobre el pool de moderador.
	incidentsSvc := incidents.NewService(moderatorQueries, []byte(hmacSecret))
	// F4 (estado_proyecto.md 2026-09-09): interest_link_categories/
	// interest_links son solo lectura para usbi_app, CRUD completo para
	// usbi_moderador; suggestions es solo INSERT para usbi_app, SELECT+DELETE
	// para usbi_moderador (00_roles_unificado.sql) — mismo split
	// PlayerService/AdminService que levels/quiz.
	interestLinksPlayerSvc := interestlinks.NewPlayerService(playerQueries)
	interestLinksAdminSvc := interestlinks.NewAdminService(moderatorQueries)
	suggestionsPlayerSvc := suggestions.NewPlayerService(playerQueries)
	suggestionsAdminSvc := suggestions.NewAdminService(moderatorQueries)
	if config.GetBoolEnv("LEGAL_MAINTENANCE_ENABLED", false) {
		// internal/maintenance no se evaluó en esta pasada de F3 (pedido
		// explícito del usuario: "el maintenance no lo toco") — se deja sobre
		// el pool de jugador sin verificar si eso es lo correcto. Está
		// deshabilitado por defecto (LEGAL_MAINTENANCE_ENABLED=false), así
		// que no es una superficie viva hoy, pero decidir su pool queda
		// pendiente para cuando se retome.
		maintenanceSvc := maintenance.NewService(playerQueries, maintenance.Config{
			InactiveSuspendAfter: config.GetDurationEnv("INACTIVE_SUSPEND_AFTER", 365*24*time.Hour),
			SuspendedCancelAfter: config.GetDurationEnv("SUSPENDED_CANCEL_AFTER", 30*24*time.Hour),
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
	// keeps level_attempts/daily_streak supplied with future yearly
	// partitions so inserts never hit the DEFAULT partition in practice.
	// Pool propio (usbi_dbmaint): sin GRANT de tabla, solo EXECUTE sobre
	// ensure_yearly_partition (migración 0005) — ver internal/dbmaint.
	if config.GetBoolEnv("DB_PARTITION_MAINTENANCE_ENABLED", true) {
		dbmaintSvc := dbmaint.NewService(dbmaintDB)
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
		AuthHandler:          auth.NewHandler(authSvc),
		QuizHandler:          quiz.NewHandler(quizAdminSvc),
		SyncHandler:          syncSvc.NewHandler(syncService),
		LevelsHandler:        levels.NewHandler(levelsPlayerSvc, levelsAdminSvc),
		DevicesHandler:       devices.NewHandler(devicesSvc),
		IncidentsHandler:     incidents.NewHandler(incidentsSvc),
		InterestLinksHandler: interestlinks.NewHandler(interestLinksPlayerSvc, interestLinksAdminSvc),
		SuggestionsHandler:   suggestions.NewHandler(suggestionsPlayerSvc, suggestionsAdminSvc),
		ReadyCheck:           readyCheck(playerDB, moderatorDB),
		TokenCfg:             tokenCfg,
		// jwtAuthMiddleware revalida token_version/status contra accounts en
		// cada petición autenticada — debe ser el pool de jugador:
		// usbi_moderador no tiene ningún GRANT sobre accounts
		// (00_roles_unificado.sql nunca se lo concede), así que
		// moderatorQueries ni siquiera podría ejecutar esta consulta.
		Repo:          playerQueries,
		AllowedOrigin: allowedOrigin,
		MaxBodyBytes:  int64(config.GetInt32Env("API_MAX_BODY_BYTES", 6*1024*1024)),
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

	// Serve in the background so main can wait for a shutdown signal.
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
// límites configurables por variables de entorno.
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

// readyCheck hace ping a los pools de jugador y moderador (F3, 2026-09-09) —
// dbmaint no se incluye aquí porque su ausencia no afecta la capacidad del
// servidor de atender peticiones HTTP, solo el mantenimiento de particiones
// en segundo plano.
func readyCheck(playerDB, moderatorDB *sql.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := playerDB.PingContext(ctx); err != nil {
			return errors.New("base de datos (jugador) inalcanzable")
		}
		if err := moderatorDB.PingContext(ctx); err != nil {
			return errors.New("base de datos (moderador) inalcanzable")
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
