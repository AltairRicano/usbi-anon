// Reescrito desde ../usbi/backend/internal/auth/service.go para el reparto en
// dos bases (plan/02_Backend.md §3.2/§3.5/§5). Register/Refresh/Logout/AgeUp/
// TutorConsent/Arco tocan casi exclusivamente `ident` (identityrepo.Queries,
// base de identidad). Login y Refresh además hacen upsert de la réplica no
// autoritativa `accounts` en `main` (repository.Queries, base principal) y
// leen su alias — es la única escritura cruzada a las dos bases fuera de la
// saga de cancelación, y no necesita ser atómica: si el upsert de accounts
// falla, el usuario simplemente no puede continuar y reintenta el login (la
// identidad ya quedó autenticada correctamente, no hay estado a medias que
// vigilar). ResolveArcoRequest implementa la saga de 3 pasos reanudable
// (pending → purging_main → identity_pseudonymized → resolved), delegando
// las dos fases de purga/seudonimización a internal/privacy.
package auth

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/mailer"
	"github.com/altair/usbi-anon-backend/internal/privacy"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// Sentinel errors — used by handler for correct HTTP status mapping.
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid credentials")
	ErrAccountSuspended = errors.New("account suspended or deleted")
	ErrEmailConflict    = errors.New("email already registered")
	ErrValidation       = errors.New("validation error")
	ErrPendingTutor     = errors.New("pending tutor consent")
	ErrInvalidRefresh   = errors.New("invalid refresh token")
	ErrForbidden        = errors.New("forbidden")
	ErrAuthBusy         = errors.New("authentication service is busy")
	// Tutor double opt-in (A1).
	ErrTutorTokenInvalid = errors.New("tutor consent token invalid")
	ErrTutorTokenExpired = errors.New("tutor consent token expired")
	ErrTutorTokenUsed    = errors.New("tutor consent token already used")
	ErrMailSend          = errors.New("failed to send verification email")
)

const defaultMaxConcurrentPasswordHashes = 2

// dummyPasswordHash is a precomputed Argon2id hash used to pad the "user not
// found" login path with the same CPU cost as a real password verification.
// Without this, an attacker can distinguish a registered email from an
// unregistered one purely by response latency, even though both cases return
// an identical error message.
var dummyPasswordHash string

func init() {
	h, err := crypto.HashPassword("usbi-timing-safe-dummy-password-do-not-use")
	if err != nil {
		panic("auth: failed to precompute dummy password hash: " + err.Error())
	}
	dummyPasswordHash = h
}

// Config holds all secrets and settings needed by auth.Service.
// Every field is required; zero values indicate a misconfiguration.
type Config struct {
	// EncryptionKey is the symmetric key for pgp_sym_encrypt/decrypt (pgcrypto).
	// Must be loaded from PGP_ENCRYPTION_KEY env var — never hardcoded.
	EncryptionKey string
	// BlindIndexSecret is the HMAC key for email lookup hashes.
	// Must be loaded from BLIND_INDEX_SECRET env var.
	BlindIndexSecret []byte
	// HMACSecret signs the privacy acceptance seal for No-Repudio.
	HMACSecret []byte
	// TokenConfig carries the JWT signing key and expiry duration.
	TokenConfig crypto.TokenConfig
	// MaxConcurrentPasswordHashes caps concurrent Argon2 work. Defaults to 2.
	MaxConcurrentPasswordHashes int
	// Mailer delivers the tutor verification email. Defaults to a dev LogMailer
	// (which only logs the link) when nil — acceptable for local dev, never prod.
	Mailer mailer.Mailer
	// TutorConsentVerifyURL is the absolute URL of the GET verification endpoint;
	// the raw token is appended as "?token=". e.g.
	// https://usbi.edu.mx/api/v1/auth/tutor-consent/verify
	TutorConsentVerifyURL string
	// TutorConsentTokenTTL is how long a verification token stays valid.
	// Defaults to 24h.
	TutorConsentTokenTTL time.Duration
}

// Service implements the authentication business logic. Recibe DOS Queries
// (una por base) en vez de una: ident habla con usbi_ident_db, main con
// usbi_anon_db — ver plan/02_Backend.md §3.2.
type Service struct {
	ident                 *identityrepo.Queries
	main                  *repository.Queries
	cfg                   Config
	passwordHashSlots     chan struct{}
	mailer                mailer.Mailer
	tutorConsentVerifyURL string
	tutorConsentTokenTTL  time.Duration
}

// NewService creates an auth.Service. It panics if cfg contains zero values
// for required secrets, preventing silent misconfigurations at startup.
func NewService(ident *identityrepo.Queries, main *repository.Queries, cfg Config) *Service {
	if len(cfg.BlindIndexSecret) == 0 {
		panic("auth.Config: BlindIndexSecret must not be empty")
	}
	if len(cfg.HMACSecret) == 0 {
		panic("auth.Config: HMACSecret must not be empty")
	}
	if cfg.EncryptionKey == "" {
		panic("auth.Config: EncryptionKey must not be empty")
	}
	if len(cfg.TokenConfig.Secret) == 0 {
		panic("auth.Config: TokenConfig.Secret must not be empty")
	}
	maxHashes := cfg.MaxConcurrentPasswordHashes
	if maxHashes <= 0 {
		maxHashes = defaultMaxConcurrentPasswordHashes
	}
	m := cfg.Mailer
	if m == nil {
		m = mailer.NewLogMailer(nil)
	}
	ttl := cfg.TutorConsentTokenTTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	verifyURL := cfg.TutorConsentVerifyURL
	if verifyURL == "" {
		verifyURL = "https://usbi.edu.mx/api/v1/auth/tutor-consent/verify"
	}
	return &Service{
		ident:                 ident,
		main:                  main,
		cfg:                   cfg,
		passwordHashSlots:     make(chan struct{}, maxHashes),
		mailer:                m,
		tutorConsentVerifyURL: verifyURL,
		tutorConsentTokenTTL:  ttl,
	}
}

// Register creates a new identity with Argon2id password hash and
// pgcrypto-encrypted email. No toca la base principal: la fila `accounts` se
// crea recién en el primer Login (plan/02_Backend.md §3.5) — un registro que
// nunca inicia sesión no necesita réplica ni alias.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	// ── Input validation ──────────────────────────────────────────────────────
	if err := validateRegister(req); err != nil {
		return RegisterResponse{}, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	// ── Password hashing (Argon2id) ───────────────────────────────────────────
	releaseHashSlot, err := s.acquirePasswordHashSlot()
	if err != nil {
		return RegisterResponse{}, err
	}
	defer releaseHashSlot()

	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("hashing password: %w", err)
	}

	// ── Blind index for email lookup ──────────────────────────────────────────
	emailHash := crypto.BlindIndexHMAC(normalizeIdentifier(req.Email), s.cfg.BlindIndexSecret)

	// ── Privacy acceptance cryptographic seal (No-Repudio) ───────────────────
	sealPayload := []byte(req.Email + "|" + req.PrivacyNoticeVersion)
	acceptanceHash := crypto.GenerateHMAC(sealPayload, s.cfg.HMACSecret)

	// ── Account status: minors require tutor consent first ───────────────────
	status := domain.StatusActive
	if !req.IsAdult {
		status = domain.StatusPendingTutorConsent
	}

	userID := uuid.New()

	identity, err := s.ident.CreateIdentity(ctx, identityrepo.CreateIdentityParams{
		ID:                      userID,
		EmailPlaintext:          strings.TrimSpace(req.Email),
		EmailLookupHash:         emailHash,
		PasswordHash:            passwordHash,
		TokenVersion:            1,
		IsAdult:                 req.IsAdult,
		Role:                    string(domain.RolePlayer),
		PrivacyNoticeVersion:    req.PrivacyNoticeVersion,
		PrivacyNoticeAcceptedAt: time.Now().UTC(),
		PrivacyAcceptanceHash:   acceptanceHash,
		CryptoKeyVersion:        1,
		Status:                  string(status),
		EncryptionKey:           s.cfg.EncryptionKey,
	})
	if err != nil {
		// PostgreSQL unique constraint violation (email_lookup_hash)
		if strings.Contains(err.Error(), "identities_email_lookup_active_idx") ||
			strings.Contains(err.Error(), "duplicate key") {
			return RegisterResponse{}, ErrEmailConflict
		}
		return RegisterResponse{}, fmt.Errorf("creating identity: %w", err)
	}

	resp := RegisterResponse{
		UserID:  identity.ID,
		Status:  identity.Status,
		Message: "User registered successfully",
	}
	if status == domain.StatusPendingTutorConsent {
		resp.RegistrationToken = tutorConsentRegistrationToken(identity.ID, s.cfg.HMACSecret)
	}
	return resp, nil
}

// Login validates credentials against la base de identidad y, si son
// válidas, sincroniza la réplica `accounts` en la base principal para que el
// resto del sistema (progreso, dispositivos) tenga una fila con la que
// referenciarse — ver plan/02_Backend.md §3.5.
// Uses constant-time comparison for password verification (timing-safe).
func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return LoginResponse{}, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	emailHash := crypto.BlindIndexHMAC(normalizeIdentifier(req.Email), s.cfg.BlindIndexSecret)

	identity, err := s.ident.GetIdentityByEmailHash(ctx, identityrepo.GetIdentityByEmailHashParams{
		EmailLookupHash: emailHash,
		EncryptionKey:   s.cfg.EncryptionKey,
	})
	if err != nil {
		// Pay the same Argon2id cost as a real login so response latency
		// doesn't leak whether the email is registered. Result is discarded —
		// this can never succeed since dummyPasswordHash never matches a real
		// password. Skip padding only if the hashing service is saturated
		// (rare, and preferable to blocking this error path).
		if releaseHashSlot, slotErr := s.acquirePasswordHashSlot(); slotErr == nil {
			_, _ = crypto.VerifyPassword(req.Password, dummyPasswordHash)
			releaseHashSlot()
		}
		// Return generic error — don't leak whether email exists.
		return LoginResponse{}, ErrUserNotFound
	}

	// Block suspended or deleted accounts before expensive hash check.
	if identity.Status == string(domain.StatusSuspended) ||
		identity.Status == string(domain.StatusDeleted) {
		return LoginResponse{}, ErrAccountSuspended
	}
	if identity.Status == string(domain.StatusPendingTutorConsent) {
		return LoginResponse{}, ErrPendingTutor
	}

	// Argon2id verification (constant-time).
	releaseHashSlot, err := s.acquirePasswordHashSlot()
	if err != nil {
		return LoginResponse{}, err
	}
	defer releaseHashSlot()

	ok, err := crypto.VerifyPassword(req.Password, identity.PasswordHash)
	if err != nil || !ok {
		return LoginResponse{}, ErrInvalidPassword
	}

	accessToken, err := s.generateAccessToken(identity.ID, domain.UserRole(identity.Role), int(identity.TokenVersion))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("generating token: %w", err)
	}
	refreshToken, refreshExpiresAt, err := s.issueRefreshToken(ctx, identity.ID)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("issuing refresh token: %w", err)
	}

	userDTO, err := s.syncAccountAndBuildUser(ctx, identity.ID, domain.UserRole(identity.Role), domain.UserStatus(identity.Status), identity.IsAdult, identity.CreatedAt)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("syncing account replica: %w", err)
	}

	return LoginResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		AccessTokenExpiresIn:  int64(s.cfg.TokenConfig.AccessExpiry.Seconds()),
		RefreshTokenExpiresAt: refreshExpiresAt,
		User:                  userDTO,
	}, nil
}

func (s *Service) acquirePasswordHashSlot() (func(), error) {
	select {
	case s.passwordHashSlots <- struct{}{}:
		return func() { <-s.passwordHashSlots }, nil
	default:
		return nil, ErrAuthBusy
	}
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (LoginResponse, error) {
	if strings.TrimSpace(req.RefreshToken) == "" {
		return LoginResponse{}, ErrInvalidRefresh
	}
	tokenHash := crypto.GenerateHMAC([]byte(req.RefreshToken), s.cfg.HMACSecret)
	refreshUser, err := s.ident.GetRefreshTokenUser(ctx, tokenHash)
	if err != nil {
		if identityrepo.IsNoRows(err) {
			return LoginResponse{}, ErrInvalidRefresh
		}
		return LoginResponse{}, err
	}
	if refreshUser.Status == string(domain.StatusSuspended) || refreshUser.Status == string(domain.StatusDeleted) {
		return LoginResponse{}, ErrAccountSuspended
	}
	if refreshUser.Status == string(domain.StatusPendingTutorConsent) {
		return LoginResponse{}, ErrPendingTutor
	}
	if err := s.ident.RevokeRefreshToken(ctx, refreshUser.TokenID); err != nil {
		return LoginResponse{}, err
	}

	accessToken, err := s.generateAccessToken(refreshUser.UserID, domain.UserRole(refreshUser.Role), int(refreshUser.TokenVersion))
	if err != nil {
		return LoginResponse{}, err
	}
	refreshToken, refreshExpiresAt, err := s.issueRefreshToken(ctx, refreshUser.UserID)
	if err != nil {
		return LoginResponse{}, err
	}

	userDTO, err := s.syncAccountAndBuildUser(ctx, refreshUser.UserID, domain.UserRole(refreshUser.Role), domain.UserStatus(refreshUser.Status), refreshUser.IsAdult, refreshUser.CreatedAt)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("syncing account replica: %w", err)
	}

	return LoginResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		AccessTokenExpiresIn:  int64(s.cfg.TokenConfig.AccessExpiry.Seconds()),
		RefreshTokenExpiresAt: refreshExpiresAt,
		User:                  userDTO,
	}, nil
}

// syncAccountAndBuildUser hace upsert de la réplica `accounts` en la base
// principal (sorteando un alias nuevo que solo se usará si la cuenta es de
// alta) y lee el alias resultante para construir el DTO domain.User. Es el
// único punto de escritura cruzada a las dos bases fuera de la saga de
// cancelación — ver plan/02_Backend.md §3.5.
func (s *Service) syncAccountAndBuildUser(ctx context.Context, userID uuid.UUID, role domain.UserRole, status domain.UserStatus, isAdult bool, createdAt time.Time) (domain.User, error) {
	adjectiveID, nounID, number, err := repository.RandomAlias()
	if err != nil {
		return domain.User{}, fmt.Errorf("generating alias: %w", err)
	}
	if err := s.main.UpsertAccount(ctx, repository.UpsertAccountParams{
		ID:               userID,
		Role:             string(role),
		Status:           string(status),
		IsAdult:          isAdult,
		AliasAdjectiveID: adjectiveID,
		AliasNounID:      nounID,
		AliasNumber:      number,
	}); err != nil {
		return domain.User{}, fmt.Errorf("upserting account: %w", err)
	}
	alias, err := s.main.GetAccountAlias(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("reading account alias: %w", err)
	}
	return domain.User{
		ID:           userID,
		DisplayAlias: alias,
		IsAdult:      isAdult,
		Role:         role,
		Status:       status,
		CreatedAt:    createdAt,
	}, nil
}

// Logout increments the user's token_version, invalidating all their current JWTs.
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrValidation
	}
	err := s.ident.IncrementTokenVersion(ctx, userID)
	if err != nil {
		return fmt.Errorf("incrementing token_version: %w", err)
	}
	if err := s.ident.RevokeRefreshTokensForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoking refresh tokens: %w", err)
	}
	return nil
}

// AgeUp attempts to transition a user from pending_tutor_consent to active.
// ip/userAgent identify the requester for the audit ledger (A3). Registra en
// identity_audit_log (no admin_audit_log): es una acción sobre la identidad,
// no sobre el progreso — ver internal/identityrepo/audit_queries.go.
func (s *Service) AgeUp(ctx context.Context, userID uuid.UUID, ip, userAgent string) error {
	if userID == uuid.Nil {
		return ErrValidation
	}

	tx, err := s.ident.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.ident.WithTx(tx)

	attempts, err := qtx.IncrementAgeUpAttempts(ctx, userID)
	if err != nil {
		return fmt.Errorf("incrementing age_up_attempts: %w", err)
	}

	if attempts > 3 {
		return errors.New("maximum age-up attempts exceeded")
	}

	if err := qtx.UpdateUserAdultStatus(ctx, userID); err != nil {
		return fmt.Errorf("updating user adult status: %w", err)
	}
	if err := qtx.PseudonymizeTutorConsents(ctx, identityrepo.PseudonymizeTutorConsentsParams{
		UserID:        userID,
		EncryptionKey: s.cfg.EncryptionKey,
	}); err != nil {
		return fmt.Errorf("pseudonymizing tutor consents after age-up: %w", err)
	}
	if err := qtx.LogIdentityAudit(ctx, identityrepo.IdentityAuditEntry{
		ActorID:    userID,
		Action:     "user.age_up",
		EntityType: "user",
		EntityID:   userID,
		After:      map[string]any{"is_adult": true, "age_up_attempts": attempts},
		IP:         ip,
		UserAgent:  userAgent,
	}); err != nil {
		return fmt.Errorf("logging age-up: %w", err)
	}

	return tx.Commit()
}

// SubmitTutorConsent starts the tutor double opt-in. It stores the tutor's data
// as a PENDING request with a single-use token (24h by default) and emails a
// verification link to the tutor. The minor's account is NOT activated here —
// activation happens only when the tutor clicks the link (VerifyTutorConsent),
// at which point the click's IP/user-agent become the legal consent evidence.
//
// requestedIP/userAgent describe whoever submitted the form; they are recorded
// for audit only, not as consent evidence. To avoid account enumeration and
// email-bombing, a token is issued (and an email sent) only when the referenced
// account exists and is actually pending tutor consent; callers always receive
// the same outcome regardless of whether that was the case.
func (s *Service) SubmitTutorConsent(ctx context.Context, req TutorConsentRequest, requestedIP net.IP, userAgent string) error {
	if err := validateTutorConsent(req); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}
	if requestedIP == nil {
		requestedIP = net.ParseIP("0.0.0.0")
	}

	status, err := s.ident.GetUserStatusByID(ctx, req.UserID)
	if err != nil {
		if identityrepo.IsNoRows(err) {
			return nil // Unknown account: behave identically to the success path.
		}
		return fmt.Errorf("looking up user: %w", err)
	}
	if status != string(domain.StatusPendingTutorConsent) {
		return nil // Not pending (active/suspended/deleted): nothing to do, no email.
	}

	// Require proof that the caller was present at registration for this
	// account. Without this check, anyone who learns a pending account's
	// UUID could submit their own tutor_email here, delete the real
	// pending token, and receive the verification link to activate someone
	// else's minor account under their own claimed identity. On mismatch
	// we no-op identically to the unknown-account branch above, so this
	// stays indistinguishable from the outside (no enumeration oracle).
	if !s.verifyRegistrationToken(req.RegistrationToken, req.UserID) {
		return nil
	}

	rawToken, err := generateOpaqueToken()
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}
	tokenHash := crypto.GenerateHMAC([]byte(rawToken), s.cfg.HMACSecret)
	expiresAt := time.Now().UTC().Add(s.tutorConsentTokenTTL)

	if err := s.storeTutorConsentToken(ctx, req, requestedIP, userAgent, tokenHash, expiresAt); err != nil {
		return err
	}

	return s.sendTutorVerificationEmail(ctx, req.TutorName, req.TutorEmail, rawToken)
}

// VerifyTutorConsent completes the tutor double opt-in. It validates the token
// the tutor clicked, records the click as legal consent evidence (tutor_consents
// row + verified token), and activates the minor's account. The click's IP and
// user-agent are the binding evidence — not those captured at form submission.
func (s *Service) VerifyTutorConsent(ctx context.Context, rawToken string, clickIP net.IP, userAgent string) error {
	if strings.TrimSpace(rawToken) == "" {
		return ErrTutorTokenInvalid
	}
	if clickIP == nil {
		clickIP = net.ParseIP("0.0.0.0")
	}
	tokenHash := crypto.GenerateHMAC([]byte(rawToken), s.cfg.HMACSecret)

	return s.processTutorConsentVerification(ctx, tokenHash, clickIP, userAgent)
}

// buildTutorConsentEmail renders the Spanish plain-text verification email.
func buildTutorConsentEmail(tutorName, link string, ttl time.Duration) string {
	greeting := "Estimado(a) tutor(a):"
	if tutorName != "" {
		greeting = "Estimado(a) " + tutorName + ":"
	}
	hours := strconv.Itoa(int(ttl.Hours()))
	return greeting + "\n\n" +
		"Un menor de edad le ha registrado como su tutor o tutora en la plataforma " +
		"USBI de la Universidad Veracruzana y proporcionó este correo para solicitar " +
		"su consentimiento sobre el tratamiento de sus datos personales.\n\n" +
		"Para autorizar la creación de la cuenta, abra el siguiente enlace dentro de " +
		"las próximas " + hours + " horas:\n\n" +
		link + "\n\n" +
		"Al abrir el enlace se registrarán la fecha, la hora y la dirección IP de su " +
		"confirmación como evidencia del consentimiento otorgado.\n\n" +
		"Si usted no reconoce esta solicitud, ignore este mensaje: la cuenta no se " +
		"activará y el enlace caducará automáticamente.\n\n" +
		"— Plataforma USBI, Universidad Veracruzana"
}

// SubmitArcoRequest records an ARCO request from the user. MarkUserDevicesForWipe
// se llama contra `main` (base principal): devices vive ahí, no en identidad.
func (s *Service) SubmitArcoRequest(ctx context.Context, userID uuid.UUID, req ArcoRequestDTO) (uuid.UUID, error) {
	if userID == uuid.Nil {
		return uuid.Nil, ErrValidation
	}
	if !isValidArcoRequestType(req.RequestType) || len(strings.TrimSpace(req.Details)) > 1000 {
		return uuid.Nil, ErrValidation
	}

	// Create the cryptographic seal for No-Repudio
	payload := []byte(userID.String() + "|" + string(req.RequestType) + "|" + req.Details)
	evidenceHash := crypto.GenerateHMAC(payload, s.cfg.HMACSecret)
	requestID := uuid.New()

	// user_id is nullable so the record survives pseudonymization after deletion.
	err := s.ident.InsertArcoRequest(ctx, identityrepo.InsertArcoRequestParams{
		ID:            requestID,
		UserID:        uuid.NullUUID{UUID: userID, Valid: true},
		RequesterType: "user",
		RequestType:   string(req.RequestType),
		Status:        "pending",
		EvidenceHash:  evidenceHash,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("inserting arco request: %w", err)
	}
	if req.RequestType == domain.ArcoCancelacion {
		if err := s.main.MarkUserDevicesForWipe(ctx, userID); err != nil {
			return uuid.Nil, fmt.Errorf("marking devices for wipe: %w", err)
		}
	}

	return requestID, nil
}

func isValidArcoRequestType(requestType domain.ArcoRequestType) bool {
	switch requestType {
	case domain.ArcoAcceso, domain.ArcoRectificacion, domain.ArcoCancelacion, domain.ArcoOposicion:
		return true
	default:
		return false
	}
}

func (s *Service) ListPendingArcoRequests(ctx context.Context, actor domain.JWTClaims, limit int32) (ArcoPendingListDTO, error) {
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleDirector {
		return ArcoPendingListDTO{}, ErrForbidden
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.ident.ListPendingArcoRequests(ctx, limit)
	if err != nil {
		return ArcoPendingListDTO{}, err
	}
	items := make([]ArcoPendingItemDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, ArcoPendingItemDTO{
			ID:            row.ID,
			RequesterType: row.RequesterType,
			RequestType:   row.RequestType,
			Status:        row.Status,
			ReceivedAt:    row.ReceivedAt,
		})
	}
	return ArcoPendingListDTO{Items: items}, nil
}

// ResolveArcoRequest implementa la saga reanudable de 3 pasos descrita en
// plan/02_Backend.md §5: pending → purging_main → identity_pseudonymized →
// resolved/rejected. El primer paso (leer + reclamar el trámite) ocurre bajo
// un SELECT ... FOR UPDATE dentro de una transacción corta de la base de
// identidad, así que dos llamadas concurrentes sobre el mismo requestID se
// serializan ahí — sin ese wrapper el FOR UPDATE sería un no-op, porque una
// consulta suelta fuera de una transacción libera cualquier lock apenas
// termina de ejecutarse. El resto de la saga (purga en [P], seudonimización
// en [I]) queda fuera de esa transacción a propósito: son operaciones en
// bases distintas y cada una es idempotente por construcción, así que un
// reintento posterior (de un segundo llamador que estaba bloqueado en el
// lock, o del job de reconciliación) nunca corrompe nada, solo repite trabajo
// ya hecho.
func (s *Service) ResolveArcoRequest(ctx context.Context, actor domain.JWTClaims, requestID uuid.UUID, req ResolveArcoRequestDTO, ip, userAgent string) error {
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleDirector {
		return ErrForbidden
	}
	if requestID == uuid.Nil || strings.TrimSpace(req.ResponseSummary) == "" {
		return ErrValidation
	}

	status := "rejected"
	if req.Approved {
		status = "resolved"
	}

	arcoReq, finalized, err := s.claimArcoRequest(ctx, actor, requestID, req, status, ip, userAgent)
	if err != nil {
		return err
	}
	if finalized {
		// No era una cancelación aprobada: no hay nada fuera de la base de
		// identidad que tocar, así que claimArcoRequest ya dejó todo resuelto.
		return nil
	}

	// A partir de aquí, req.Approved && isCancelacion está garantizado —
	// claimArcoRequest solo devuelve finalized=false en ese caso. El propio
	// checkpoint (arcoReq.Status) le dice a ResumeArcoCancellation por dónde
	// retomar: handled_by/response_summary ya quedaron guardados en el
	// reclamo (identityrepo.ClaimArcoRequestForCancellation), así que solo
	// falta ejecutar las fases pendientes y cerrar el trámite.
	if err := privacy.ResumeArcoCancellation(ctx, s.ident, s.main, requestID, arcoReq.Status, privacy.CancelParams{
		UserID:           arcoReq.UserID.UUID,
		Reason:           "arco_cancelacion",
		EncryptionKey:    s.cfg.EncryptionKey,
		BlindIndexSecret: s.cfg.BlindIndexSecret,
	}); err != nil {
		return fmt.Errorf("resuming arco cancellation: %w", err)
	}

	// No-Repudio: registra la decisión del admin ya reflejada en la base
	// (A3). No usa finalizeArcoRequest/ResolveArcoRequest de nuevo: el
	// estado terminal (status/resolved_at/handled_by/response_summary) ya lo
	// dejó ResumeArcoCancellation vía MarkArcoRequestResolved + el reclamo
	// previo; aquí solo falta la entrada de auditoría.
	if err := s.ident.LogIdentityAudit(ctx, identityrepo.IdentityAuditEntry{
		ActorID:    actor.UserID,
		Action:     "arco.resolve",
		EntityType: "arco_request",
		EntityID:   requestID,
		Before:     map[string]any{"status": "pending", "request_type": arcoReq.RequestType},
		After:      map[string]any{"status": status, "approved": req.Approved, "subject_user_id": arcoReq.UserID.UUID},
		IP:         ip,
		UserAgent:  userAgent,
	}); err != nil {
		return fmt.Errorf("logging arco resolution: %w", err)
	}
	return nil
}

// claimArcoRequest bloquea la fila de arco_requests (FOR UPDATE) dentro de
// una transacción corta y decide, atómicamente respecto a cualquier llamada
// concurrente sobre el mismo requestID, qué sigue:
//
//   - Si NO es una cancelación aprobada (rechazo, o cualquier otro tipo de
//     trámite ARCO): no hay nada que tocar en la base principal, así que se
//     resuelve y audita en esta misma transacción y finalized=true.
//   - Si SÍ es una cancelación aprobada y el trámite está en 'pending': se
//     reclama el primer checkpoint ('purging_main') y se libera el lock —
//     desde aquí el resto de la saga corre fuera de esta transacción.
//   - Si ya estaba en 'purging_main' o 'identity_pseudonymized' (un intento
//     previo murió a mitad de camino, o esta llamada estaba bloqueada en el
//     lock detrás de otra que ya avanzó el checkpoint): se retoma desde ahí.
func (s *Service) claimArcoRequest(ctx context.Context, actor domain.JWTClaims, requestID uuid.UUID, req ResolveArcoRequestDTO, status, ip, userAgent string) (identityrepo.ArcoRequestForResolution, bool, error) {
	tx, err := s.ident.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return identityrepo.ArcoRequestForResolution{}, false, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.ident.WithTx(tx)

	arcoReq, err := qtx.GetArcoRequestForUpdate(ctx, requestID)
	if err != nil {
		return arcoReq, false, err
	}
	switch arcoReq.Status {
	case "pending", "purging_main", "identity_pseudonymized":
		// continúa abajo
	default:
		return arcoReq, false, ErrValidation
	}

	isCancelacion := arcoReq.RequestType == string(domain.ArcoCancelacion) && arcoReq.UserID.Valid
	if !(req.Approved && isCancelacion) {
		if err := s.finalizeArcoRequest(ctx, qtx, actor, requestID, status, arcoReq, req, ip, userAgent); err != nil {
			return arcoReq, false, err
		}
		return arcoReq, true, tx.Commit()
	}

	if arcoReq.Status == "pending" {
		if err := qtx.ClaimArcoRequestForCancellation(ctx, requestID,
			uuid.NullUUID{UUID: actor.UserID, Valid: true},
			strings.TrimSpace(req.ResponseSummary),
		); err != nil {
			return arcoReq, false, fmt.Errorf("claiming purging_main checkpoint: %w", err)
		}
		arcoReq.Status = "purging_main"
	}
	return arcoReq, false, tx.Commit()
}

// finalizeArcoRequest escribe la decisión final (resolved/rejected) y su
// entrada de auditoría de No-Repudio (A3). Se usa desde dos sitios: el
// atajo de claimArcoRequest (trámites sin trabajo cruzado a bases) y el
// cierre normal de la saga completa en ResolveArcoRequest.
func (s *Service) finalizeArcoRequest(ctx context.Context, qtx *identityrepo.Queries, actor domain.JWTClaims, requestID uuid.UUID, status string, arcoReq identityrepo.ArcoRequestForResolution, req ResolveArcoRequestDTO, ip, userAgent string) error {
	if err := qtx.ResolveArcoRequest(ctx, identityrepo.ResolveArcoRequestParams{
		ID:              requestID,
		HandledBy:       uuid.NullUUID{UUID: actor.UserID, Valid: true},
		Status:          status,
		ResponseSummary: strings.TrimSpace(req.ResponseSummary),
	}); err != nil {
		return fmt.Errorf("resolving arco request: %w", err)
	}

	var subjectID uuid.UUID
	if arcoReq.UserID.Valid {
		subjectID = arcoReq.UserID.UUID
	}
	if err := qtx.LogIdentityAudit(ctx, identityrepo.IdentityAuditEntry{
		ActorID:    actor.UserID,
		Action:     "arco.resolve",
		EntityType: "arco_request",
		EntityID:   requestID,
		Before:     map[string]any{"status": "pending", "request_type": arcoReq.RequestType},
		After:      map[string]any{"status": status, "approved": req.Approved, "subject_user_id": subjectID},
		IP:         ip,
		UserAgent:  userAgent,
	}); err != nil {
		return fmt.Errorf("logging arco resolution: %w", err)
	}
	return nil
}

// ── Validation helpers ────────────────────────────────────────────────────────

func validateRegister(req RegisterRequest) error {
	var errs []string
	if !strings.Contains(req.Email, "@") || len(req.Email) < 5 {
		errs = append(errs, "email is invalid")
	}
	if len(req.Password) < 8 {
		errs = append(errs, "password must be at least 8 characters")
	}
	if req.PrivacyNoticeVersion == "" {
		errs = append(errs, "privacy_notice_version is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func validateLogin(req LoginRequest) error {
	var errs []string
	if !strings.Contains(req.Email, "@") || len(req.Email) < 5 {
		errs = append(errs, "email is invalid")
	}
	if req.Password == "" {
		errs = append(errs, "password is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (s *Service) generateAccessToken(userID uuid.UUID, role domain.UserRole, tokenVersion int) (string, error) {
	claims := domain.JWTClaims{
		UserID:       userID,
		Role:         role,
		TokenVersion: tokenVersion,
	}
	return crypto.GenerateToken(claims, s.cfg.TokenConfig)
}

func (s *Service) issueRefreshToken(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	token, err := generateOpaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	tokenHash := crypto.GenerateHMAC([]byte(token), s.cfg.HMACSecret)
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	if err := s.ident.InsertRefreshToken(ctx, identityrepo.InsertRefreshTokenParams{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// generateOpaqueToken returns a URL-safe, 256-bit random token. Used for both
// refresh tokens and tutor-consent verification links.
// tutorConsentRegistrationToken derives a deterministic, verifiable proof that
// the caller was present at registration time for userID. It is handed to the
// client in RegisterResponse and must be echoed back in SubmitTutorConsent —
// without it, anyone who merely learns a pending account's UUID could submit
// themselves as the tutor and hijack the double opt-in (see audit finding).
func tutorConsentRegistrationToken(userID uuid.UUID, secret []byte) string {
	mac := crypto.GenerateHMAC([]byte("tutor-consent-registration|"+userID.String()), secret)
	return base64.RawURLEncoding.EncodeToString(mac)
}

func generateOpaqueToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := cryptorand.Read(tokenBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func normalizeIdentifier(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateTutorConsent(req TutorConsentRequest) error {
	var errs []string
	if req.UserID == uuid.Nil {
		errs = append(errs, "user_id is required")
	}
	if strings.TrimSpace(req.TutorName) == "" {
		errs = append(errs, "tutor_name is required")
	}
	if !strings.Contains(req.TutorEmail, "@") || len(req.TutorEmail) < 5 {
		errs = append(errs, "tutor_email is invalid")
	}
	if req.PrivacyNoticeVersion == "" {
		errs = append(errs, "privacy_notice_version is required")
	}
	if strings.TrimSpace(req.RegistrationToken) == "" {
		errs = append(errs, "registration_token is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// ── Private helpers for tutor consent to improve functional cohesion ──────────

func (s *Service) verifyRegistrationToken(registrationToken string, userID uuid.UUID) bool {
	presentedMAC, decodeErr := base64.RawURLEncoding.DecodeString(registrationToken)
	if decodeErr != nil {
		return false
	}
	return crypto.VerifyHMAC([]byte("tutor-consent-registration|"+userID.String()), presentedMAC, s.cfg.HMACSecret)
}

func (s *Service) storeTutorConsentToken(ctx context.Context, req TutorConsentRequest, requestedIP net.IP, userAgent string, tokenHash []byte, expiresAt time.Time) error {
	tx, err := s.ident.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.ident.WithTx(tx)

	if err := qtx.DeleteUnverifiedTutorConsentTokensForUser(ctx, req.UserID); err != nil {
		return fmt.Errorf("clearing previous tutor consent tokens: %w", err)
	}
	if err := qtx.InsertTutorConsentToken(ctx, identityrepo.InsertTutorConsentTokenParams{
		ID:                   uuid.New(),
		UserID:               req.UserID,
		TutorName:            strings.TrimSpace(req.TutorName),
		TutorEmail:           strings.ToLower(strings.TrimSpace(req.TutorEmail)),
		PrivacyNoticeVersion: req.PrivacyNoticeVersion,
		TokenHash:            tokenHash,
		RequestedIP:          requestedIP,
		RequestedUserAgent:   userAgent,
		CryptoKeyVersion:     1,
		ExpiresAt:            expiresAt,
		EncryptionKey:        s.cfg.EncryptionKey,
	}); err != nil {
		return fmt.Errorf("inserting tutor consent token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing tutor consent token: %w", err)
	}
	return nil
}

func (s *Service) sendTutorVerificationEmail(ctx context.Context, tutorName, tutorEmail, rawToken string) error {
	tutorEmail = strings.ToLower(strings.TrimSpace(tutorEmail))
	link := s.tutorConsentVerifyURL + "?token=" + rawToken
	subject := "Verificación de consentimiento — Plataforma USBI"
	body := buildTutorConsentEmail(strings.TrimSpace(tutorName), link, s.tutorConsentTokenTTL)
	if err := s.mailer.Send(ctx, tutorEmail, subject, body); err != nil {
		return fmt.Errorf("%w: %v", ErrMailSend, err)
	}
	return nil
}

func (s *Service) processTutorConsentVerification(ctx context.Context, tokenHash []byte, clickIP net.IP, userAgent string) error {
	tx, err := s.ident.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.ident.WithTx(tx)

	row, err := qtx.GetTutorConsentTokenByHashForUpdate(ctx, tokenHash, s.cfg.EncryptionKey)
	if err != nil {
		if identityrepo.IsNoRows(err) {
			return ErrTutorTokenInvalid
		}
		return fmt.Errorf("looking up tutor consent token: %w", err)
	}
	if row.VerifiedAt.Valid {
		return ErrTutorTokenUsed
	}
	if time.Now().UTC().After(row.ExpiresAt) {
		return ErrTutorTokenExpired
	}

	tutorEmail := strings.ToLower(strings.TrimSpace(row.TutorEmail))
	signaturePayload := []byte(row.UserID.String() + "|" + tutorEmail + "|" + row.PrivacyNoticeVersion)
	signature := crypto.GenerateHMAC(signaturePayload, s.cfg.HMACSecret)
	now := time.Now().UTC()

	if err := qtx.InsertTutorConsent(ctx, identityrepo.InsertTutorConsentParams{
		ID:                   uuid.New(),
		UserID:               row.UserID,
		TutorName:            strings.TrimSpace(row.TutorName),
		TutorEmail:           tutorEmail,
		PrivacyNoticeVersion: row.PrivacyNoticeVersion,
		AcceptedAt:           now,
		AcceptanceIP:         clickIP,
		AcceptanceUserAgent:  userAgent,
		ConsentSignature:     signature,
		CryptoKeyVersion:     1,
		EncryptionKey:        s.cfg.EncryptionKey,
	}); err != nil {
		return fmt.Errorf("inserting tutor consent: %w", err)
	}
	if err := qtx.MarkTutorConsentTokenVerified(ctx, identityrepo.MarkTutorConsentTokenVerifiedParams{
		ID:                    row.ID,
		VerifiedAt:            now,
		VerificationIP:        clickIP,
		VerificationUserAgent: userAgent,
	}); err != nil {
		return fmt.Errorf("marking tutor consent token verified: %w", err)
	}
	if err := qtx.ActivateTutorConsentUser(ctx, row.UserID); err != nil {
		return fmt.Errorf("activating tutor consent user: %w", err)
	}
	return tx.Commit()
}
