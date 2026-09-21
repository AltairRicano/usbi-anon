package levels

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler mantiene una referencia a cada uno de los dos Service partidos en
// PlayerService y AdminService usan pools de base de datos distintos.ivos de un rol, el Handler solo
// habla con su Service correspondiente; para ListLevels/GetLevel/ListSections
// —que sirven a ambos perfiles— decide con cuál hablar usando la misma señal
// de rol que ya existía antes del split (canManageContent).
type Handler struct {
	player *PlayerService
	admin  *AdminService
}

func NewHandler(player *PlayerService, admin *AdminService) *Handler {
	return &Handler{player: player, admin: admin}
}

func (h *Handler) CreateLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can create levels")
		return
	}

	var req CreateLevelRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.admin.CreateLevel(r.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
		} else {
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not create level")
		}
		return
	}

	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

// ListLevels sirve a ambos perfiles: quien pueda administrar contenido puede
// pedir include_unpublished=true y ve el catálogo completo con el pool de
// usbi_moderador; cualquier otra petición autenticada solo ve lo publicado,
// con el pool de usbi_app. La decisión de qué Service invocar vive aquí (ya
// existía antes del split); qué datos devuelve cada uno vive en su propio
// Service — ver PlayerService.ListLevels / AdminService.ListLevels.
func (h *Handler) ListLevels(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	claims, _ := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)

	cursor := uuid.Nil
	if c := q.Get("cursor"); c != "" {
		parsed, err := uuid.Parse(c)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "cursor must be a valid UUID")
			return
		}
		cursor = parsed
	}

	pageSize := int32(20)
	if ps := q.Get("page_size"); ps != "" {
		n, err := strconv.Atoi(ps)
		if err != nil || n < 1 || n > 50 {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "page_size must be between 1 and 50")
			return
		}
		pageSize = int32(n)
	}

	sectionID := uuid.Nil
	if rawSectionID := q.Get("section_id"); rawSectionID != "" {
		parsed, err := uuid.Parse(rawSectionID)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "section_id must be a valid UUID")
			return
		}
		sectionID = parsed
	}

	var (
		page LevelsPage
		err  error
	)
	if claims != nil && canManageContent(claims.Role) {
		includeUnpublished := q.Get("include_unpublished") == "true"
		page, err = h.admin.ListLevels(r.Context(), cursor, sectionID, includeUnpublished, pageSize)
	} else {
		page, err = h.player.ListLevels(r.Context(), cursor, sectionID, pageSize)
	}
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not retrieve levels")
		return
	}

	httpproblem.WriteJSON(w, http.StatusOK, page)
}

// GetLevel: mismo criterio de despacho que ListLevels. A diferencia de
// ListLevels, ni PlayerService ni AdminService exponen un parámetro
// includeUnpublished — cada uno siempre se comporta según su rol (ver
// comentarios en player_service.go / admin_service.go).
func (h *Handler) GetLevel(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	var (
		resp LevelResponse
		err  error
	)
	if claims != nil && canManageContent(claims.Role) {
		resp, err = h.admin.GetLevel(r.Context(), levelID)
	} else {
		resp, err = h.player.GetLevel(r.Context(), levelID)
	}
	if err != nil {
		writeServiceError(w, r, err, "Could not retrieve level")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can update levels")
		return
	}

	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	var req UpdateLevelRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.admin.UpdateLevel(r.Context(), claims.UserID, levelID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not update level")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PublishLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can publish levels")
		return
	}

	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	resp, err := h.admin.PublishLevel(r.Context(), claims.UserID, levelID)
	if err != nil {
		writeServiceError(w, r, err, "Could not publish level")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) UnpublishLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can unpublish levels")
		return
	}

	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	resp, err := h.admin.UnpublishLevel(r.Context(), claims.UserID, levelID)
	if err != nil {
		writeServiceError(w, r, err, "Could not unpublish level")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ArchiveLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canArchiveContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins or directors can archive levels")
		return
	}

	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	resp, err := h.admin.ArchiveLevel(r.Context(), claims.UserID, levelID)
	if err != nil {
		writeServiceError(w, r, err, "Could not archive level")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListArchivedLevels(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can list archived levels")
		return
	}

	sectionID := uuid.Nil
	if raw := r.URL.Query().Get("section_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "section_id must be a valid UUID")
			return
		}
		sectionID = parsed
	}

	resp, err := h.admin.ListArchivedLevels(r.Context(), sectionID)
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not retrieve archived levels")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) UnarchiveLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canArchiveContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can restore archived levels")
		return
	}

	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	resp, err := h.admin.UnarchiveLevel(r.Context(), claims.UserID, levelID)
	if err != nil {
		writeServiceError(w, r, err, "Could not restore level")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PurgeLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canArchiveContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can purge levels")
		return
	}

	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	if err := h.admin.PurgeLevel(r.Context(), claims.UserID, levelID); err != nil {
		writeServiceError(w, r, err, "Could not purge level")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CompleteLevel(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Missing JWT claims in context")
		return
	}

	levelID, ok := parseURLUUID(w, r, "level_id")
	if !ok {
		return
	}

	var req CompleteLevelRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.player.CompleteLevel(r.Context(), claims.UserID, levelID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not complete level")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetProfileProgress(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Missing JWT claims in context")
		return
	}

	resp, err := h.player.GetProfileProgress(r.Context(), claims.UserID)
	if err != nil {
		writeServiceError(w, r, err, "Could not retrieve progress")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can create sections")
		return
	}

	var req CreateSectionRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.admin.CreateSection(r.Context(), claims.UserID, req)
	if err != nil {
		log.Printf("CreateSection error: %v", err)
		writeServiceError(w, r, err, "Could not create section")
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

// ListSections: mismo criterio de despacho que ListLevels/GetLevel.
func (h *Handler) ListSections(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)

	var (
		resp SectionsResponse
		err  error
	)
	if claims != nil && canManageContent(claims.Role) {
		includeUnpublished := r.URL.Query().Get("include_unpublished") == "true"
		resp, err = h.admin.ListSections(r.Context(), includeUnpublished)
	} else {
		resp, err = h.player.ListSections(r.Context())
	}
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not retrieve sections")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can update sections")
		return
	}

	sectionID, ok := parseURLUUID(w, r, "section_id")
	if !ok {
		return
	}

	var req UpdateSectionRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.admin.UpdateSection(r.Context(), claims.UserID, sectionID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not update section")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PublishSection(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can publish sections")
		return
	}

	sectionID, ok := parseURLUUID(w, r, "section_id")
	if !ok {
		return
	}

	resp, err := h.admin.PublishSection(r.Context(), claims.UserID, sectionID)
	if err != nil {
		writeServiceError(w, r, err, "Could not publish section")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) UnpublishSection(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can unpublish sections")
		return
	}

	sectionID, ok := parseURLUUID(w, r, "section_id")
	if !ok {
		return
	}

	resp, err := h.admin.UnpublishSection(r.Context(), claims.UserID, sectionID)
	if err != nil {
		writeServiceError(w, r, err, "Could not unpublish section")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ArchiveSection(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canArchiveContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins or directors can archive sections")
		return
	}

	sectionID, ok := parseURLUUID(w, r, "section_id")
	if !ok {
		return
	}

	resp, err := h.admin.ArchiveSection(r.Context(), claims.UserID, sectionID)
	if err != nil {
		writeServiceError(w, r, err, "Could not archive section")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListArchivedSections(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canManageContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only content managers can list archived sections")
		return
	}

	resp, err := h.admin.ListArchivedSections(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not retrieve archived sections")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) UnarchiveSection(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canArchiveContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can restore archived sections")
		return
	}

	sectionID, ok := parseURLUUID(w, r, "section_id")
	if !ok {
		return
	}

	resp, err := h.admin.UnarchiveSection(r.Context(), claims.UserID, sectionID)
	if err != nil {
		writeServiceError(w, r, err, "Could not restore section")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PurgeSection(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canArchiveContent(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can purge sections")
		return
	}

	sectionID, ok := parseURLUUID(w, r, "section_id")
	if !ok {
		return
	}

	if err := h.admin.PurgeSection(r.Context(), claims.UserID, sectionID); err != nil {
		writeServiceError(w, r, err, "Could not purge section")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseURLUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", key+" must be a valid UUID")
		return uuid.Nil, false
	}
	return parsed, true
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, ErrValidation):
		httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
	case errors.Is(err, ErrNotFound):
		httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "Resource not found")
	case errors.Is(err, ErrForbidden):
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", err.Error())
	case errors.Is(err, ErrNotArchived):
		httpproblem.WriteProblem(w, r, http.StatusConflict, "not-archived", "Conflict", "Content must be archived before it can be purged or restored")
	case errors.Is(err, ErrSectionHasLevels):
		httpproblem.WriteProblem(w, r, http.StatusConflict, "section-has-levels", "Conflict", "Purge every level in this section (archived or not) before purging the section itself")
	default:
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", fallback)
	}
}

func canManageContent(role domain.UserRole) bool {
	return role == domain.RoleAdmin
}

func canArchiveContent(role domain.UserRole) bool {
	return role == domain.RoleAdmin
}
