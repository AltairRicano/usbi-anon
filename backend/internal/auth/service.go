// Reescrito en F9 (plan/04_Rediseno_identidad_gustos.md §2 y §3) para el
// esquema unificado sin email ni flujo de tutor. El diseño anterior (dos
// bases, correo cifrado, doble opt-in de tutor por correo) queda descartado
// por completo: no hay Register de una sola llamada, sino tres pasos
// (RegisterQuestions → RegisterAnswers → RegisterConfirm, ver
// registration_token.go), y no hay SubmitTutorConsent/VerifyTutorConsent —
// un menor autoreportado juega de inmediato.
package auth

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/privacy"
	"github.com/altair/usbi-anon-backend/internal/quiz"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Sentinel errors — used by handler for correct HTTP status mapping.
var (
	ErrValidation           = errors.New("validation error")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidPassword      = errors.New("invalid credentials")
	ErrAccountSuspended     = errors.New("account suspended or deleted")
	ErrNicknameConflict     = errors.New("nickname already taken")
	ErrInvalidRefresh       = errors.New("invalid refresh token")
	ErrForbidden            = errors.New("forbidden")
	ErrAuthBusy             = errors.New("authentication service is busy")
	ErrNotFound             = errors.New("not found")
	ErrCannotDeleteAdmin    = errors.New("cannot delete an admin account")
	ErrTooManyAgeUpAttempts = errors.New("maximum age-up attempts exceeded")
)

const defaultMaxConcurrentPasswordHashes = 2

// dummyPasswordHash is a precomputed Argon2id hash used to pad the
// "nickname not found" login path with the same CPU cost as a real password
// verification, so response latency doesn't leak whether a nickname exists.
var dummyPasswordHash string

func init() {
	h, err := crypto.HashPassword("usbi-timing-safe-dummy-password-do-not-use")
	if err != nil {
		panic("auth: failed to precompute dummy password hash: " + err.Error())
	}
	dummyPasswordHash = h
}

// Config holds all secrets and settings needed by auth.Service.
type Config struct {
	// HMACSecret firma el token de registro y el sello de aceptación del
	// aviso de privacidad, y los tokens de refresh.
	HMACSecret []byte
	// TokenConfig carries the JWT signing key and expiry duration.
	TokenConfig crypto.TokenConfig
	// MaxConcurrentPasswordHashes caps concurrent Argon2 work. Defaults to 2.
	MaxConcurrentPasswordHashes int
	// StaffPrivacyNoticeVersion se sella en las cuentas creadas por un admin
	// (POST /admin/accounts, decisión 6 del rediseño) — esas cuentas no
	// pasan por el cuestionario de gustos, así que no traen su propia
	// versión del aviso. Placeholder hasta la reescritura legal completa
	// (F11): no hay todavía un aviso de privacidad específico para staff.
	StaffPrivacyNoticeVersion string
}

// Service implementa la lógica de autenticación contra la única base del
// sistema. quiz orquesta la generación de nickname/password y el banco de
// preguntas — auth nunca genera candidatos ni toca registration_questions
// directo, siempre a través de ese paquete.
type Service struct {
	repo              *repository.Queries
	quiz              *quiz.Service
	cfg               Config
	passwordHashSlots chan struct{}
}

// NewService creates an auth.Service. It panics if cfg contains zero values
// for required secrets, preventing silent misconfigurations at startup.
func NewService(repo *repository.Queries, quizSvc *quiz.Service, cfg Config) *Service {
	if len(cfg.HMACSecret) == 0 {
		panic("auth.Config: HMACSecret must not be empty")
	}
	if len(cfg.TokenConfig.Secret) == 0 {
		panic("auth.Config: TokenConfig.Secret must not be empty")
	}
	maxHashes := cfg.MaxConcurrentPasswordHashes
	if maxHashes <= 0 {
		maxHashes = defaultMaxConcurrentPasswordHashes
	}
	if cfg.StaffPrivacyNoticeVersion == "" {
		cfg.StaffPrivacyNoticeVersion = "staff-bootstrap-v1"
	}
	return &Service{
		repo:              repo,
		quiz:              quizSvc,
		cfg:               cfg,
		passwordHashSlots: make(chan struct{}, maxHashes),
	}
}

// ── Registro en 3 pasos ──────────────────────────────────────────────────

// RegisterQuestions handles the first step: a random subset of active
// questions. Delegates entirely to internal/quiz — auth no decide el
// muestreo, solo traduce el DTO.
func (s *Service) RegisterQuestions(ctx context.Context) (RegisterQuestionsResponse, error) {
	resp, err := s.quiz.SelectQuestionsForRegistration(ctx)
	if err != nil {
		return RegisterQuestionsResponse{}, err
	}
	questions := make([]QuestionOption, 0, len(resp.Questions))
	for _, q := range resp.Questions {
		questions = append(questions, QuestionOption{ID: q.ID, Text: q.Text})
	}
	return RegisterQuestionsResponse{Questions: questions, MaxQuestionsShown: resp.MaxQuestionsShown}, nil
}

// RegisterAnswers valida las respuestas, genera 4 candidatos de nickname y
// devuelve el estado firmado que RegisterConfirm necesitará — sin escribir
// nada en la base todavía: una persona que abandona aquí no deja ninguna
// fila a medias.
func (s *Service) RegisterAnswers(ctx context.Context, req RegisterAnswersRequest) (RegisterAnswersResponse, error) {
	if strings.TrimSpace(req.PrivacyNoticeVersion) == "" {
		return RegisterAnswersResponse{}, fmt.Errorf("%w: privacy_notice_version is required", ErrValidation)
	}
	if len(req.Answers) < 2 || len(req.Answers) > 10 {
		return RegisterAnswersResponse{}, fmt.Errorf("%w: answers must contain between 2 and 10 items", ErrValidation)
	}

	answerTexts := make([]string, 0, len(req.Answers))
	payloadAnswers := make([]answerPayload, 0, len(req.Answers))
	seenQuestions := make(map[uuid.UUID]bool, len(req.Answers))

	for _, a := range req.Answers {
		text := strings.TrimSpace(a.AnswerText)
		if err := validateAnswerText(text); err != nil {
			return RegisterAnswersResponse{}, fmt.Errorf("%w: %s", ErrValidation, err.Error())
		}
		if seenQuestions[a.QuestionID] {
			return RegisterAnswersResponse{}, fmt.Errorf("%w: duplicate question_id", ErrValidation)
		}
		seenQuestions[a.QuestionID] = true

		question, err := s.quiz.GetActiveQuestionByID(ctx, a.QuestionID)
		if err != nil {
			if errors.Is(err, quiz.ErrNotFound) {
				return RegisterAnswersResponse{}, fmt.Errorf("%w: unknown or inactive question_id", ErrValidation)
			}
			return RegisterAnswersResponse{}, err
		}

		answerTexts = append(answerTexts, text)
		payloadAnswers = append(payloadAnswers, answerPayload{
			QuestionID:           a.QuestionID,
			QuestionTextSnapshot: question.QuestionText,
			AnswerText:           text,
		})
	}

	candidates, err := quiz.GenerateNicknameCandidates(answerTexts, s.nicknameExists(ctx))
	if err != nil {
		if errors.Is(err, quiz.ErrInsufficientAnswers) {
			return RegisterAnswersResponse{}, fmt.Errorf("%w: answers do not contain enough usable characters to derive a nickname", ErrValidation)
		}
		return RegisterAnswersResponse{}, err
	}

	token, err := s.signRegistrationToken(registrationTokenPayload{
		Answers:              payloadAnswers,
		IsAdult:              req.IsAdult,
		PrivacyNoticeVersion: strings.TrimSpace(req.PrivacyNoticeVersion),
		NicknameCandidates:   candidates[:],
		ExpiresAt:            time.Now().UTC().Add(registrationTokenTTL),
	})
	if err != nil {
		return RegisterAnswersResponse{}, fmt.Errorf("signing registration token: %w", err)
	}

	return RegisterAnswersResponse{RegistrationToken: token, NicknameCandidates: candidates[:]}, nil
}

// RegisterConfirm valida el nickname elegido contra los 4 candidatos
// firmados, RE-verifica la colisión (pudo tomarse en los minutos que pasaron
// desde /answers), genera el password y crea la cuenta — respuestas,
// account_quiz_answers y alias en una sola transacción.
func (s *Service) RegisterConfirm(ctx context.Context, req RegisterConfirmRequest) (RegisterConfirmResponse, error) {
	payload, err := s.verifyRegistrationToken(req.RegistrationToken)
	if err != nil {
		return RegisterConfirmResponse{}, err
	}

	chosen := strings.TrimSpace(req.ChosenNickname)
	valid := false
	for _, c := range payload.NicknameCandidates {
		if c == chosen {
			valid = true
			break
		}
	}
	if !valid {
		return RegisterConfirmResponse{}, fmt.Errorf("%w: chosen_nickname is not one of the offered candidates", ErrValidation)
	}

	exists, err := s.nicknameExists(ctx)(chosen)
	if err != nil {
		return RegisterConfirmResponse{}, err
	}
	if exists {
		return RegisterConfirmResponse{}, ErrNicknameConflict
	}

	answerTexts := make([]string, 0, len(payload.Answers))
	for _, a := range payload.Answers {
		answerTexts = append(answerTexts, a.AnswerText)
	}
	plainPassword, err := quiz.GeneratePassword(answerTexts)
	if err != nil {
		return RegisterConfirmResponse{}, fmt.Errorf("generating password: %w", err)
	}

	releaseHashSlot, err := s.acquirePasswordHashSlot()
	if err != nil {
		return RegisterConfirmResponse{}, err
	}
	defer releaseHashSlot()
	passwordHash, err := crypto.HashPassword(plainPassword)
	if err != nil {
		return RegisterConfirmResponse{}, fmt.Errorf("hashing password: %w", err)
	}

	adjectiveID, nounID, number, err := repository.RandomAlias()
	if err != nil {
		return RegisterConfirmResponse{}, fmt.Errorf("generating alias: %w", err)
	}

	accountID := uuid.New()
	acceptedAt := time.Now().UTC()
	acceptanceHash := crypto.GenerateHMAC(privacyAcceptanceSealPayload(accountID, payload.PrivacyNoticeVersion, acceptedAt), s.cfg.HMACSecret)

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return RegisterConfirmResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	account, err := qtx.CreateAccount(ctx, repository.CreateAccountParams{
		ID:                      accountID,
		Nickname:                chosen,
		PasswordHash:            passwordHash,
		IsAdult:                 payload.IsAdult,
		Role:                    string(domain.RolePlayer),
		AliasAdjectiveID:        adjectiveID,
		AliasNounID:             nounID,
		AliasNumber:             number,
		PrivacyNoticeVersion:    payload.PrivacyNoticeVersion,
		PrivacyNoticeAcceptedAt: acceptedAt,
		PrivacyAcceptanceHash:   acceptanceHash,
		CryptoKeyVersion:        1,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return RegisterConfirmResponse{}, ErrNicknameConflict
		}
		return RegisterConfirmResponse{}, fmt.Errorf("creating account: %w", err)
	}

	for _, a := range payload.Answers {
		if err := qtx.InsertAccountQuizAnswer(ctx, repository.InsertAccountQuizAnswerParams{
			ID:                   uuid.New(),
			AccountID:            accountID,
			QuestionID:           a.QuestionID,
			QuestionTextSnapshot: a.QuestionTextSnapshot,
			AnswerText:           a.AnswerText,
		}); err != nil {
			return RegisterConfirmResponse{}, fmt.Errorf("inserting quiz answer: %w", err)
		}
	}

	alias, err := qtx.GetAccountAlias(ctx, accountID)
	if err != nil {
		return RegisterConfirmResponse{}, fmt.Errorf("reading alias: %w", err)
	}

	if err := logAuditEntry(ctx, qtx, accountID, "account.register", "account", accountID, nil,
		map[string]any{"nickname": chosen}, "", ""); err != nil {
		return RegisterConfirmResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return RegisterConfirmResponse{}, err
	}

	return RegisterConfirmResponse{
		AccountID:    accountID,
		Nickname:     account.Nickname,
		Password:     plainPassword,
		DisplayAlias: alias,
	}, nil
}

func (s *Service) nicknameExists(ctx context.Context) func(string) (bool, error) {
	return func(nickname string) (bool, error) {
		_, err := s.repo.FindAccountByNickname(ctx, nickname)
		if err != nil {
			if repository.IsNoRows(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
}

func privacyAcceptanceSealPayload(accountID uuid.UUID, version string, acceptedAt time.Time) []byte {
	return []byte(accountID.String() + "|" + version + "|" + acceptedAt.Format(time.RFC3339Nano))
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// ── Login / sesión ────────────────────────────────────────────────────────

// Login busca la cuenta directo por nickname (sin blind index: no es PII
// cifrada) y verifica el password con comparación en tiempo constante.
func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return LoginResponse{}, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	account, err := s.repo.FindAccountByNickname(ctx, normalizeNickname(req.Nickname))
	if err != nil {
		// Paga el mismo costo de Argon2id que un login real para que la
		// latencia no filtre si el nickname existe. El resultado se
		// descarta — nunca puede tener éxito, dummyPasswordHash no coincide
		// con ningún password real.
		if releaseHashSlot, slotErr := s.acquirePasswordHashSlot(); slotErr == nil {
			_, _ = crypto.VerifyPassword(req.Password, dummyPasswordHash)
			releaseHashSlot()
		}
		return LoginResponse{}, ErrUserNotFound
	}

	if account.Status == string(domain.StatusSuspended) || account.Status == string(domain.StatusDeleted) {
		return LoginResponse{}, ErrAccountSuspended
	}

	releaseHashSlot, err := s.acquirePasswordHashSlot()
	if err != nil {
		return LoginResponse{}, err
	}
	defer releaseHashSlot()
	ok, err := crypto.VerifyPassword(req.Password, account.PasswordHash)
	if err != nil || !ok {
		return LoginResponse{}, ErrInvalidPassword
	}

	return s.issueSession(ctx, account)
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
	rt, err := s.repo.GetRefreshTokenAccount(ctx, tokenHash)
	if err != nil {
		if repository.IsNoRows(err) {
			return LoginResponse{}, ErrInvalidRefresh
		}
		return LoginResponse{}, err
	}
	if rt.Status == string(domain.StatusSuspended) || rt.Status == string(domain.StatusDeleted) {
		return LoginResponse{}, ErrAccountSuspended
	}
	if err := s.repo.RevokeRefreshToken(ctx, rt.TokenID); err != nil {
		return LoginResponse{}, err
	}

	account, err := s.repo.GetAccountByID(ctx, rt.AccountID)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("reading account: %w", err)
	}
	return s.issueSession(ctx, account)
}

// issueSession emite el par access/refresh, registra el instante de
// actividad y compone el domain.User que ve el cliente. Login y Refresh
// terminan aquí para no duplicar los cuatro pasos.
func (s *Service) issueSession(ctx context.Context, account repository.Account) (LoginResponse, error) {
	accessToken, err := s.generateAccessToken(account.ID, domain.UserRole(account.Role), int(account.TokenVersion))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("generating token: %w", err)
	}
	refreshToken, refreshExpiresAt, err := s.issueRefreshToken(ctx, account.ID)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("issuing refresh token: %w", err)
	}
	if err := s.repo.TouchAccountLastLogin(ctx, account.ID); err != nil {
		return LoginResponse{}, fmt.Errorf("touching last login: %w", err)
	}
	alias, err := s.repo.GetAccountAlias(ctx, account.ID)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("reading alias: %w", err)
	}

	return LoginResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		AccessTokenExpiresIn:  int64(s.cfg.TokenConfig.AccessExpiry.Seconds()),
		RefreshTokenExpiresAt: refreshExpiresAt,
		User: domain.User{
			ID:           account.ID,
			Nickname:     account.Nickname,
			DisplayAlias: alias,
			IsAdult:      account.IsAdult,
			Role:         domain.UserRole(account.Role),
			Status:       domain.UserStatus(account.Status),
			CreatedAt:    account.CreatedAt,
		},
	}, nil
}

// Logout increments the account's token_version, invalidating all its
// current JWTs, and revokes every live refresh token.
func (s *Service) Logout(ctx context.Context, accountID uuid.UUID) error {
	if accountID == uuid.Nil {
		return ErrValidation
	}
	if err := s.repo.IncrementAccountTokenVersion(ctx, accountID); err != nil {
		return fmt.Errorf("incrementing token_version: %w", err)
	}
	if err := s.repo.RevokeRefreshTokensForAccount(ctx, accountID); err != nil {
		return fmt.Errorf("revoking refresh tokens: %w", err)
	}
	return nil
}

// AgeUp aplica la transición a mayoría de edad (Ley 251, máx. 3 intentos).
// Simplificado frente al diseño de dos bases: sin tutor que pseudonimizar,
// sin identity_audit_log — solo el contador y el estado, auditados en la
// misma bitácora unificada que cualquier otra acción.
func (s *Service) AgeUp(ctx context.Context, accountID uuid.UUID, ip, userAgent string) error {
	if accountID == uuid.Nil {
		return ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	attempts, err := qtx.IncrementAgeUpAttempts(ctx, accountID)
	if err != nil {
		return fmt.Errorf("incrementing age_up_attempts: %w", err)
	}
	if attempts > 3 {
		return ErrTooManyAgeUpAttempts
	}
	if err := qtx.MarkAccountAdult(ctx, accountID); err != nil {
		return fmt.Errorf("marking account adult: %w", err)
	}
	if err := logAuditEntry(ctx, qtx, accountID, "account.age_up", "account", accountID, nil,
		map[string]any{"is_adult": true, "age_up_attempts": attempts}, ip, userAgent); err != nil {
		return fmt.Errorf("logging age-up: %w", err)
	}

	return tx.Commit()
}

// CancelSelf implementa DELETE /auth/me: cancelación autoservicio inmediata
// (decisión 7 del rediseño), sin aprobación de nadie. Una sola llamada a
// internal/privacy.CancelAccount, ya transaccional.
func (s *Service) CancelSelf(ctx context.Context, accountID uuid.UUID) error {
	if accountID == uuid.Nil {
		return ErrValidation
	}
	return privacy.CancelAccount(ctx, s.repo, privacy.CancelAccountParams{
		AccountID: accountID,
		Reason:    "self_service_cancellation",
	})
}

// ── ARCO (acceso / rectificación / oposición) ────────────────────────────

// SubmitArcoRequest registra una solicitud ARCO. "cancelacion" se rechaza a
// propósito: esa vía ahora es DELETE /auth/me, autoservicio e inmediata —
// dejarla aceptable aquí resucitaría la pregunta de qué hace un admin con
// una "cancelación" que ya no tiene ninguna acción pendiente de aprobar.
func (s *Service) SubmitArcoRequest(ctx context.Context, accountID uuid.UUID, req ArcoRequestDTO) (uuid.UUID, error) {
	if accountID == uuid.Nil {
		return uuid.Nil, ErrValidation
	}
	if req.RequestType == domain.ArcoCancelacion {
		return uuid.Nil, fmt.Errorf("%w: use DELETE /auth/me to cancel your account, not this endpoint", ErrValidation)
	}
	if !isValidArcoRequestType(req.RequestType) || len(strings.TrimSpace(req.Details)) > 1000 {
		return uuid.Nil, ErrValidation
	}

	payload := []byte(accountID.String() + "|" + string(req.RequestType) + "|" + req.Details)
	evidenceHash := crypto.GenerateHMAC(payload, s.cfg.HMACSecret)
	requestID := uuid.New()

	if err := s.repo.InsertArcoRequest(ctx, repository.InsertArcoRequestParams{
		ID:            requestID,
		UserID:        uuid.NullUUID{UUID: accountID, Valid: true},
		RequesterType: "user",
		RequestType:   string(req.RequestType),
		EvidenceHash:  evidenceHash,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("inserting arco request: %w", err)
	}
	return requestID, nil
}

func isValidArcoRequestType(t domain.ArcoRequestType) bool {
	switch t {
	case domain.ArcoAcceso, domain.ArcoRectificacion, domain.ArcoOposicion:
		return true
	default:
		return false
	}
}

func (s *Service) ListPendingArcoRequests(ctx context.Context, actor domain.JWTClaims, limit int32) (ArcoPendingListDTO, error) {
	if actor.Role != domain.RoleAdmin {
		return ArcoPendingListDTO{}, ErrForbidden
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.repo.ListPendingArcoRequests(ctx, limit)
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

// ResolveArcoRequest ya no orquesta ninguna saga (plan/04 §1.2 y §2): con
// cancelacion fuera de este flujo, resolver un trámite es un solo UPDATE
// bajo lock, dentro de una transacción corta que solo existe para que dos
// admins no puedan resolver el mismo trámite a la vez.
func (s *Service) ResolveArcoRequest(ctx context.Context, actor domain.JWTClaims, requestID uuid.UUID, req ResolveArcoRequestDTO, ip, userAgent string) error {
	if actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	if requestID == uuid.Nil || strings.TrimSpace(req.ResponseSummary) == "" {
		return ErrValidation
	}

	status := "rejected"
	if req.Approved {
		status = "resolved"
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	arcoReq, err := qtx.GetArcoRequestForUpdate(ctx, requestID)
	if err != nil {
		if repository.IsNoRows(err) {
			return ErrNotFound
		}
		return err
	}
	if arcoReq.Status != "pending" {
		return ErrValidation
	}

	if err := qtx.ResolveArcoRequest(ctx, repository.ResolveArcoRequestParams{
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
	if err := logAuditEntry(ctx, qtx, actor.UserID, "arco.resolve", "arco_request", requestID,
		map[string]any{"status": "pending", "request_type": arcoReq.RequestType},
		map[string]any{"status": status, "approved": req.Approved, "subject_account_id": subjectID},
		ip, userAgent,
	); err != nil {
		return fmt.Errorf("logging arco resolution: %w", err)
	}

	return tx.Commit()
}

// ── Administración de cuentas ────────────────────────────────────────────

// CreateAdminAccount da de alta staff (admin/operator/director) o incluso un
// player adicional, con nickname+password explícitos — sin cuestionario de
// gustos (decisión 6 del rediseño).
func (s *Service) CreateAdminAccount(ctx context.Context, actor domain.JWTClaims, req AdminCreateAccountRequest) (AdminAccountResponse, error) {
	if actor.Role != domain.RoleAdmin {
		return AdminAccountResponse{}, ErrForbidden
	}
	nickname := normalizeNickname(req.Nickname)
	if !nicknameFormatValid(nickname) {
		return AdminAccountResponse{}, fmt.Errorf("%w: nickname must be 6-20 lowercase letters/digits", ErrValidation)
	}
	if len(req.Password) < adminPasswordMinLen {
		return AdminAccountResponse{}, fmt.Errorf("%w: password must be at least %d characters", ErrValidation, adminPasswordMinLen)
	}
	if !isValidStaffRole(req.Role) {
		return AdminAccountResponse{}, fmt.Errorf("%w: invalid role", ErrValidation)
	}

	releaseHashSlot, err := s.acquirePasswordHashSlot()
	if err != nil {
		return AdminAccountResponse{}, err
	}
	defer releaseHashSlot()
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return AdminAccountResponse{}, fmt.Errorf("hashing password: %w", err)
	}

	adjectiveID, nounID, number, err := repository.RandomAlias()
	if err != nil {
		return AdminAccountResponse{}, fmt.Errorf("generating alias: %w", err)
	}

	accountID := uuid.New()
	acceptedAt := time.Now().UTC()
	acceptanceHash := crypto.GenerateHMAC(privacyAcceptanceSealPayload(accountID, s.cfg.StaffPrivacyNoticeVersion, acceptedAt), s.cfg.HMACSecret)

	account, err := s.repo.CreateAccount(ctx, repository.CreateAccountParams{
		ID:                      accountID,
		Nickname:                nickname,
		PasswordHash:            passwordHash,
		IsAdult:                 true,
		Role:                    string(req.Role),
		AliasAdjectiveID:        adjectiveID,
		AliasNounID:             nounID,
		AliasNumber:             number,
		PrivacyNoticeVersion:    s.cfg.StaffPrivacyNoticeVersion,
		PrivacyNoticeAcceptedAt: acceptedAt,
		PrivacyAcceptanceHash:   acceptanceHash,
		CryptoKeyVersion:        1,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return AdminAccountResponse{}, ErrNicknameConflict
		}
		return AdminAccountResponse{}, fmt.Errorf("creating account: %w", err)
	}

	alias, err := s.repo.GetAccountAlias(ctx, accountID)
	if err != nil {
		return AdminAccountResponse{}, fmt.Errorf("reading alias: %w", err)
	}

	if err := logAuditEntry(ctx, s.repo, actor.UserID, "admin_account.create", "account", accountID, nil,
		map[string]any{"nickname": nickname, "role": string(req.Role)}, "", ""); err != nil {
		return AdminAccountResponse{}, err
	}

	return AdminAccountResponse{
		ID:           account.ID,
		Nickname:     account.Nickname,
		Role:         domain.UserRole(account.Role),
		DisplayAlias: alias,
		CreatedAt:    account.CreatedAt,
	}, nil
}

func isValidStaffRole(role domain.UserRole) bool {
	switch role {
	case domain.RoleAdmin, domain.RolePlayer:
		return true
	default:
		return false
	}
}

// DeleteAccount es el borrado administrativo: mismo mecanismo que la
// cancelación autoservicio (internal/privacy.CancelAccount), disparado por
// un admin en vez de por la propia persona usuaria. Un admin NUNCA puede
// borrar a otro admin — sin excepción (§2 del rediseño): si hace falta,
// primero hay que degradar el rol de la cuenta objetivo.
func (s *Service) DeleteAccount(ctx context.Context, actor domain.JWTClaims, targetID uuid.UUID) error {
	if actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	if targetID == uuid.Nil {
		return ErrValidation
	}
	target, err := s.repo.GetAccountByID(ctx, targetID)
	if err != nil {
		if repository.IsNoRows(err) {
			return ErrNotFound
		}
		return err
	}
	if target.Role == string(domain.RoleAdmin) {
		return ErrCannotDeleteAdmin
	}
	if err := privacy.CancelAccount(ctx, s.repo, privacy.CancelAccountParams{
		AccountID: targetID,
		Reason:    "admin_deletion",
	}); err != nil {
		return err
	}
	return logAuditEntry(ctx, s.repo, actor.UserID, "admin_account.delete", "account", targetID, nil, nil, "", "")
}

// GetAccountQuizAnswers expone las respuestas del cuestionario para que un
// admin pueda comparar y decidir si resetea el password de una cuenta
// olvidada (§1 decisión 4). Auditado en cada lectura: es el único endpoint
// de todo el sistema que devuelve texto libre escrito por una persona
// usuaria.
func (s *Service) GetAccountQuizAnswers(ctx context.Context, actor domain.JWTClaims, targetID uuid.UUID) (AdminQuizAnswersResponse, error) {
	if actor.Role != domain.RoleAdmin {
		return AdminQuizAnswersResponse{}, ErrForbidden
	}
	rows, err := s.repo.ListAccountQuizAnswers(ctx, targetID)
	if err != nil {
		return AdminQuizAnswersResponse{}, err
	}
	items := make([]AdminQuizAnswerDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, AdminQuizAnswerDTO{
			QuestionTextSnapshot: r.QuestionTextSnapshot,
			AnswerText:           r.AnswerText,
			CreatedAt:            r.CreatedAt,
		})
	}
	if err := logAuditEntry(ctx, s.repo, actor.UserID, "admin_account.view_quiz_answers", "account", targetID, nil,
		map[string]any{"answer_count": len(items)}, "", ""); err != nil {
		return AdminQuizAnswersResponse{}, err
	}
	return AdminQuizAnswersResponse{Items: items}, nil
}

// ResetAccountPassword genera un password aleatorio nuevo (no derivado del
// cuestionario: un reseteo administrativo no tiene por qué reutilizar ese
// algoritmo) y lo hashea. El texto plano se devuelve una única vez, igual
// que en el registro.
func (s *Service) ResetAccountPassword(ctx context.Context, actor domain.JWTClaims, targetID uuid.UUID) (AdminResetPasswordResponse, error) {
	if actor.Role != domain.RoleAdmin {
		return AdminResetPasswordResponse{}, ErrForbidden
	}
	if targetID == uuid.Nil {
		return AdminResetPasswordResponse{}, ErrValidation
	}

	newPassword, err := generateStaffPassword()
	if err != nil {
		return AdminResetPasswordResponse{}, fmt.Errorf("generating password: %w", err)
	}

	releaseHashSlot, err := s.acquirePasswordHashSlot()
	if err != nil {
		return AdminResetPasswordResponse{}, err
	}
	defer releaseHashSlot()
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return AdminResetPasswordResponse{}, fmt.Errorf("hashing password: %w", err)
	}

	if err := s.repo.SetAccountPassword(ctx, targetID, passwordHash); err != nil {
		return AdminResetPasswordResponse{}, fmt.Errorf("setting password: %w", err)
	}
	if err := logAuditEntry(ctx, s.repo, actor.UserID, "admin_account.reset_password", "account", targetID, nil, nil, "", ""); err != nil {
		return AdminResetPasswordResponse{}, err
	}

	return AdminResetPasswordResponse{NewPassword: newPassword}, nil
}

const (
	staffPasswordLength  = 16
	staffPasswordCharset = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789" // sin 0/O/1/l/I
)

func generateStaffPassword() (string, error) {
	b := make([]byte, staffPasswordLength)
	for i := range b {
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(staffPasswordCharset))))
		if err != nil {
			return "", err
		}
		b[i] = staffPasswordCharset[n.Int64()]
	}
	return string(b), nil
}

// ── Helpers de token ──────────────────────────────────────────────────────

func (s *Service) generateAccessToken(accountID uuid.UUID, role domain.UserRole, tokenVersion int) (string, error) {
	claims := domain.JWTClaims{
		UserID:       accountID,
		Role:         role,
		TokenVersion: tokenVersion,
	}
	return crypto.GenerateToken(claims, s.cfg.TokenConfig)
}

func (s *Service) issueRefreshToken(ctx context.Context, accountID uuid.UUID) (string, time.Time, error) {
	token, err := generateOpaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	tokenHash := crypto.GenerateHMAC([]byte(token), s.cfg.HMACSecret)
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	if err := s.repo.InsertRefreshToken(ctx, repository.InsertRefreshTokenParams{
		ID:        uuid.New(),
		AccountID: accountID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// generateOpaqueToken returns a URL-safe, 256-bit random token used for
// refresh tokens.
func generateOpaqueToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := cryptorand.Read(tokenBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func logAuditEntry(ctx context.Context, repo *repository.Queries, actorID uuid.UUID, action, entityType string, entityID uuid.UUID, before, after any, ip, userAgent string) error {
	return audit.Log(ctx, repo, audit.Entry{
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Before:     before,
		After:      after,
		IP:         ip,
		UserAgent:  userAgent,
	})
}
