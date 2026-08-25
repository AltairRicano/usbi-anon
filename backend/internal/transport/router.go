package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/altair/usbi-anon-backend/internal/auth"
	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/devices"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/altair/usbi-anon-backend/internal/httputil"
	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/incidents"
	"github.com/altair/usbi-anon-backend/internal/levels"
	syncHandler "github.com/altair/usbi-anon-backend/internal/sync"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterDependencies holds all handler and config dependencies. A nil handler
// registers a 501 "not implemented" stub for its routes instead of panicking,
// so partial wiring (e.g. in tests) is safe.
type RouterDependencies struct {
	AuthHandler      *auth.Handler
	SyncHandler      *syncHandler.Handler
	LevelsHandler    *levels.Handler
	DevicesHandler   *devices.Handler
	IncidentsHandler *incidents.Handler
	ReadyCheck       func(context.Context) error
	TokenCfg         crypto.TokenConfig
	// IdentQueries valida token_version contra la BASE DE IDENTIDAD — la
	// única de las dos que sabe si un usuario cerró sesión o fue revocado.
	// A diferencia de ../usbi, aquí NO puede ser *repository.Queries: esa
	// tabla no vive en la base principal.
	IdentQueries *identityrepo.Queries
	// MaxBodyBytes caps incoming API request bodies. Defaults to 6 MiB.
	MaxBodyBytes int64
	// AllowedOrigin is a comma-separated CORS allowlist (or "*" to opt into
	// allowing any origin). Defaults to "https://usbi.edu.mx" if empty.
	AllowedOrigin string
	// TrustProxyHeaders enables chi's RealIP middleware, which rewrites
	// r.RemoteAddr from True-Client-IP/X-Real-IP/X-Forwarded-For headers.
	// MUST stay false unless a reverse proxy in front of this service is
	// confirmed to strip/set those headers itself — otherwise any direct
	// client can spoof its own IP for rate limiting and audit logging.
	TrustProxyHeaders bool
	// RequestTimeout bounds every request via middleware.Timeout. Defaults to 20s.
	RequestTimeout time.Duration
}

const defaultMaxBodyBytes int64 = 6 * 1024 * 1024

// ClaimsFromContext extracts the JWT claims injected by jwtAuthMiddleware.
// Returns nil if no claims are present (public route or missing middleware).
func ClaimsFromContext(ctx context.Context) *domain.JWTClaims {
	if v := ctx.Value(domain.ClaimsKey); v != nil {
		if c, ok := v.(*domain.JWTClaims); ok {
			return c
		}
	}
	return nil
}

// SetupRoutes registers all HTTP routes with their middleware chain. It
// returns a cleanup func that stops the rate limiters' background cleanup
// goroutine — callers should defer it (or invoke it during graceful
// shutdown); it is not required for the server to function correctly.
func SetupRoutes(r chi.Router, deps RouterDependencies) func() {
	rl := newRateLimiters()

	origin := deps.AllowedOrigin
	if origin == "" {
		origin = "https://usbi.edu.mx"
	}

	requestTimeout := deps.RequestTimeout
	if requestTimeout <= 0 {
		requestTimeout = 20 * time.Second
	}

	// Global middleware stack
	r.Use(middleware.RequestID)
	if deps.TrustProxyHeaders {
		r.Use(middleware.RealIP)
	}
	r.Use(requestLogger)
	r.Use(middleware.Recoverer)
	// Per-request deadline (B5): cancels the request context so slow queries /
	// blocked advisory locks release their goroutine and pool connection.
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(securityHeaders)
	r.Use(corsMiddleware(origin))
	r.Use(maxBodyBytesMiddleware(deps.MaxBodyBytes))

	r.Route("/api/v1", func(r chi.Router) {
		// ── Public routes ─────────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(rl.authMiddleware)
			if deps.AuthHandler != nil {
				r.Post("/auth/register", deps.AuthHandler.Register)
				r.Post("/auth/login", deps.AuthHandler.Login)
				r.Post("/auth/refresh", deps.AuthHandler.Refresh)
				r.Post("/auth/tutor-consent", deps.AuthHandler.TutorConsent)
				r.Get("/auth/tutor-consent/verify", deps.AuthHandler.VerifyTutorConsent)
			} else {
				r.Post("/auth/register", notImplementedHandler("auth.register"))
				r.Post("/auth/login", notImplementedHandler("auth.login"))
				r.Post("/auth/refresh", notImplementedHandler("auth.refresh"))
				r.Post("/auth/tutor-consent", notImplementedHandler("auth.tutorConsent"))
				r.Get("/auth/tutor-consent/verify", notImplementedHandler("auth.verifyTutorConsent"))
			}
		})

		// ── Authenticated routes ──────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(jwtAuthMiddleware(deps.TokenCfg, deps.IdentQueries))
			r.Use(rl.generalMiddleware)

			if deps.AuthHandler != nil {
				r.Post("/auth/logout", deps.AuthHandler.Logout)
				r.Post("/auth/age-up", deps.AuthHandler.AgeUp)
				r.Post("/arco", deps.AuthHandler.Arco)
				r.Get("/arco/pending", deps.AuthHandler.ListPendingArco)
				r.Post("/arco/{request_id}/resolve", deps.AuthHandler.ResolveArco)
			} else {
				r.Post("/auth/logout", notImplementedHandler("auth.logout"))
				r.Post("/auth/age-up", notImplementedHandler("auth.ageUp"))
				r.Post("/arco", notImplementedHandler("arco.submitRequest"))
				r.Get("/arco/pending", notImplementedHandler("arco.listPending"))
				r.Post("/arco/{request_id}/resolve", notImplementedHandler("arco.resolveRequest"))
			}

			if deps.SyncHandler != nil {
				r.Post("/sync", deps.SyncHandler.SyncData)
			} else {
				r.Post("/sync", notImplementedHandler("sync.offlineProgress"))
			}

			if deps.IncidentsHandler != nil {
				r.Post("/admin/security-incidents", deps.IncidentsHandler.CreateIncident)
			} else {
				r.Post("/admin/security-incidents", notImplementedHandler("admin.createSecurityIncident"))
			}

			if deps.DevicesHandler != nil {
				r.Post("/devices", deps.DevicesHandler.RegisterDevice)
				r.Get("/devices", deps.DevicesHandler.ListDevices)
			} else {
				r.Post("/devices", notImplementedHandler("devices.register"))
				r.Get("/devices", notImplementedHandler("devices.list"))
			}

			// Level routes (Phase 4 — Maker module)
			if deps.LevelsHandler != nil {
				r.Get("/sections", deps.LevelsHandler.ListSections)
				r.Post("/sections", deps.LevelsHandler.CreateSection)
				r.Patch("/sections/{section_id}", deps.LevelsHandler.UpdateSection)
				r.Post("/sections/{section_id}/publish", deps.LevelsHandler.PublishSection)
				r.Post("/sections/{section_id}/unpublish", deps.LevelsHandler.UnpublishSection)
				r.Post("/sections/{section_id}/archive", deps.LevelsHandler.ArchiveSection)

				r.Post("/levels", deps.LevelsHandler.CreateLevel)
				r.Get("/levels", deps.LevelsHandler.ListLevels)
				r.Get("/levels/{level_id}", deps.LevelsHandler.GetLevel)
				r.Patch("/levels/{level_id}", deps.LevelsHandler.UpdateLevel)
				r.Post("/levels/{level_id}/publish", deps.LevelsHandler.PublishLevel)
				r.Post("/levels/{level_id}/unpublish", deps.LevelsHandler.UnpublishLevel)
				r.Post("/levels/{level_id}/archive", deps.LevelsHandler.ArchiveLevel)
				r.Post("/levels/{level_id}/complete", deps.LevelsHandler.CompleteLevel)

				r.Get("/profile/progress", deps.LevelsHandler.GetProfileProgress)
			} else {
				r.Get("/sections", notImplementedHandler("sections.list"))
				r.Post("/sections", notImplementedHandler("sections.create"))
				r.Post("/levels", notImplementedHandler("levels.create"))
				r.Get("/levels", notImplementedHandler("levels.list"))
				r.Get("/profile/progress", notImplementedHandler("profile.progress"))
			}
		})
	})

	// Health check — unauthenticated, unversioned
	r.Get("/health", healthHandler)
	r.Get("/health/live", healthHandler)
	r.Get("/health/ready", readyHandler(deps.ReadyCheck))

	return rl.Close
}

func maxBodyBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBodyBytes
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// requestLogger emits one structured slog record per request — method, path,
// status, bytes, latency, request id and client IP (B4). It replaces chi's
// text-only middleware.Logger so logs are machine-readable, and its status +
// duration fields double as the minimal request metrics (rate-limit rejections
// show up as status=429, HMAC failures as the /sync error status, etc.).
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
			"remote_ip", httputil.ClientIP(r),
		)
	})
}

// healthHandler returns a simple liveness probe response.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"usbi-anon-backend"}`))
}

// readyHandler expone en el body cuál dependencia falló (criterio 5 de
// plan/02_Backend.md §8: con dos bases independientes, "la base de datos no
// responde" ya no es suficiente detalle para operar el servicio — hace falta
// saber si es la principal o la de identidad). check (ver readyCheck en
// main.go) ya se encarga de no filtrar el error crudo del driver.
func readyHandler(check func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check == nil {
			healthHandler(w, r)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := check(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			body, _ := json.Marshal(map[string]string{
				"status":  "unavailable",
				"service": "usbi-anon-backend",
				"detail":  err.Error(),
			})
			_, _ = w.Write(body)
			return
		}

		healthHandler(w, r)
	}
}

// notImplementedHandler returns a 501 stub for routes pending implementation.
func notImplementedHandler(operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpproblem.WriteProblem(w, r, http.StatusNotImplemented, "not-implemented",
			"Not Implemented",
			"Operation '"+operation+"' is pending implementation.")
	}
}

// securityHeaders sets conservative, framework-agnostic security response
// headers on every response (audit finding CN-003). It hardens the JSON API and
// any direct browser navigation to an API URL — including the tutor-consent
// verify link, whose Referrer-Policy: no-referrer stops the single-use token
// from leaking via the Referer header (CN-006). The SPA served by the reverse
// proxy still needs its own Content-Security-Policy configured there.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		// Only assert HSTS when this process terminates TLS itself; behind a
		// reverse proxy, the TLS terminator (Nginx) owns this header.
		if r.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware applies CORS headers. The origin is configurable to support
// both production (https://usbi.edu.mx) and local LAN development.
// corsMiddleware enforces an allowlist parsed from a comma-separated list of
// origins (e.g. "https://usbi.edu.mx,http://localhost:5173"). A literal "*"
// entry opts into allowing any origin (discouraged outside local development).
// Unlike a naive reflect-any-origin implementation, a request whose Origin
// does not match gets no Access-Control-Allow-Origin header at all, so
// browsers block the cross-origin response as intended.
func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := make(map[string]struct{})
	wildcard := false
	for _, o := range strings.Split(allowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			wildcard = true
			continue
		}
		origins[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")

			reqOrigin := r.Header.Get("Origin")
			if reqOrigin != "" {
				if wildcard {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if _, ok := origins[reqOrigin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// jwtAuthMiddleware validates the JWT and injects claims into the request context.
// Downstream handlers retrieve claims via ClaimsFromContext(r.Context()).
// identQueries valida token_version contra la BASE DE IDENTIDAD (identities),
// no contra la principal — ../usbi usaba *repository.Queries porque ahí
// token_version vivía junto con el resto de `users`.
func jwtAuthMiddleware(cfg crypto.TokenConfig, identQueries *identityrepo.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
					"Unauthorized", "Missing or malformed Authorization header")
				return
			}

			tokenStr := authHeader[7:]

			// Reject if JWT secret is not configured (zero-value cfg).
			if len(cfg.Secret) == 0 {
				httpproblem.WriteProblem(w, r, http.StatusServiceUnavailable, "misconfigured",
					"Service Unavailable", "Authentication service is not configured")
				return
			}

			claims, err := crypto.ValidateToken(tokenStr, cfg)
			if err != nil {
				httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
					"Unauthorized", "Invalid or expired token")
				return
			}

			// Verify token_version against the database if identQueries is provided.
			// This enables immediate token revocation on logout/password change.
			if identQueries != nil {
				dbVersion, err := identQueries.GetUserTokenVersion(r.Context(), claims.UserID)
				if err != nil || int(dbVersion) != claims.TokenVersion {
					httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
						"Unauthorized", "Token has been revoked or is invalid")
					return
				}
			}

			// Inject claims into context for downstream handlers.
			ctx := context.WithValue(r.Context(), domain.ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Stricter limiter for public auth routes (register/login/refresh/tutor-consent),
// where the cost of a request (Argon2id hashing) makes even modest per-IP rates
// attractive for credential stuffing / brute force. ~1 attempt every 2s with a
// burst of 5 comfortably covers a legitimate user mistyping a password.
const (
	authRateRPS   rate.Limit = 0.5
	authRateBurst            = 5
)

// generalRateRPS/generalRateBurst govern the general authenticated-route limiter.
const (
	generalRateRPS   rate.Limit = 10
	generalRateBurst            = 20
)

// proxyMisconfigWarnThreshold is how many requests must be observed before the
// single-IP-diversity check below fires (avoids a false positive from the first
// few requests at cold start).
const proxyMisconfigWarnThreshold = 50

// visitorTracker is one bucket of per-IP token-bucket limiters. It owns its
// own lock, so distinct trackers (general vs. auth routes) never contend.
type visitorTracker struct {
	mu           sync.Mutex
	visitors     map[string]*visitor
	requestCount int
	rps          rate.Limit
	burst        int
}

func newVisitorTracker(rps rate.Limit, burst int) *visitorTracker {
	return &visitorTracker{visitors: make(map[string]*visitor), rps: rps, burst: burst}
}

func (t *visitorTracker) getLimiter(ip string) *rate.Limiter {
	t.mu.Lock()
	defer t.mu.Unlock()
	v, exists := t.visitors[ip]
	if !exists {
		v = &visitor{limiter: rate.NewLimiter(t.rps, t.burst), lastSeen: time.Now()}
		t.visitors[ip] = v
	} else {
		v.lastSeen = time.Now()
	}
	t.requestCount++
	return v.limiter
}

func (t *visitorTracker) expireIdle(maxIdle time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for ip, v := range t.visitors {
		if time.Since(v.lastSeen) > maxIdle {
			delete(t.visitors, ip)
		}
	}
}

func (t *visitorTracker) snapshot() (distinctIPs, requests int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.visitors), t.requestCount
}

// rateLimiters owns both per-IP limiter buckets (general + auth) and the
// goroutine that expires idle entries. It replaces what used to be
// package-level global maps/mutexes plus an uncancellable init() goroutine —
// each RouterDependencies instance now gets its own, stoppable via Close(),
// which also makes it safe to spin up multiple independent routers (e.g. in
// tests) without sharing rate-limit state between them.
type rateLimiters struct {
	general            *visitorTracker
	auth               *visitorTracker
	proxyMisconfigOnce sync.Once
	stop               chan struct{}
}

func newRateLimiters() *rateLimiters {
	rl := &rateLimiters{
		general: newVisitorTracker(generalRateRPS, generalRateBurst),
		auth:    newVisitorTracker(authRateRPS, authRateBurst),
		stop:    make(chan struct{}),
	}
	go rl.runCleanup()
	return rl
}

// Close stops the background cleanup goroutine. Safe to call once.
func (rl *rateLimiters) Close() {
	close(rl.stop)
}

func (rl *rateLimiters) runCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.general.expireIdle(3 * time.Minute)
			rl.auth.expireIdle(3 * time.Minute)
		case <-rl.stop:
			return
		}
	}
}

// warnIfLikelyProxyMisconfigured logs once if every request so far has come
// from the same IP after a meaningful sample size — the signature of
// TRUST_PROXY_HEADERS=false behind a reverse proxy that strips/rewrites the
// client's real IP, collapsing rate limiting and tutor-consent audit IPs onto
// the proxy's loopback address (audit finding B3).
func (rl *rateLimiters) warnIfLikelyProxyMisconfigured() {
	distinctGeneral, generalCount := rl.general.snapshot()
	distinctAuth, authCount := rl.auth.snapshot()

	total := generalCount + authCount
	if total < proxyMisconfigWarnThreshold {
		return
	}
	if distinctGeneral > 1 || distinctAuth > 1 {
		return
	}
	rl.proxyMisconfigOnce.Do(func() {
		slog.Warn("rate limiter has seen only one distinct client IP after many requests; "+
			"if this server sits behind a reverse proxy, TRUST_PROXY_HEADERS is probably "+
			"false when it should be true (see DEPLOYMENT.md)",
			"requests_observed", total)
	})
}

func (rl *rateLimiters) generalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rl.general.getLimiter(httputil.ClientIP(r))
		rl.warnIfLikelyProxyMisconfigured()
		if !limiter.Allow() {
			httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "rate-limit-exceeded",
				"Too Many Requests", "Rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware applies a much stricter per-IP limit to the public
// authentication routes, which are the most attractive target for credential
// stuffing / brute force and previously had no rate limiting at all.
func (rl *rateLimiters) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rl.auth.getLimiter(httputil.ClientIP(r))
		rl.warnIfLikelyProxyMisconfigured()
		if !limiter.Allow() {
			httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "rate-limit-exceeded",
				"Too Many Requests", "Rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}
