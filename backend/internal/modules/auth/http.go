package auth

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/v2/httpx"
	"backend/internal/platform/v2/middleware"
	"backend/internal/platform/v2/validator"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type authService interface {
	Register(context.Context, RegisterDTO) (user.User, error)
	Login(context.Context, LoginDTO) (user.User, session.Session, string, error)
	Me(context.Context, string) (user.User, error)
	Logout(context.Context, string) error
}

type Handler struct {
	svc          authService
	log          *slog.Logger
	secureCookie bool
}

func NewHandler(svc authService, log *slog.Logger, secureCookie bool) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{svc: svc, log: log, secureCookie: secureCookie}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("POST /auth/logout", h.logout)
	mux.Handle("GET /auth/me", middleware.AuthGuard(h.me))
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())
	u, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusOK, u)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	if cookie, err := r.Cookie("session"); err == nil {
		if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
			h.sendError(w, r, err)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: "", Path: "/",
		HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
		Expires: time.Unix(1, 0).UTC(), MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	var payload LoginDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		h.sendError(w, r, err)
		return
	}
	if err := payload.Validate(); err != nil {
		h.sendError(w, r, err)
		return
	}
	u, created, token, err := h.svc.Login(r.Context(), payload)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: token, Path: "/",
		HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
		Expires: created.ExpiresAt, MaxAge: int(session.Lifetime.Seconds()),
	})
	h.sendSuccess(w, r, http.StatusOK, u)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var payload RegisterDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		h.sendError(w, r, err)
		return
	}
	if err := payload.Validate(); err != nil {
		h.sendError(w, r, err)
		return
	}
	u, err := h.svc.Register(r.Context(), payload)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusCreated, u)
}

func (h *Handler) sendSuccess(w http.ResponseWriter, r *http.Request, status int, data any) {
	if err := httpx.SendSuccess(w, status, httpx.NewSuccessResponse("", data)); err != nil {
		h.log.ErrorContext(r.Context(), "write auth response", "err", err)
	}
}

func (h *Handler) sendError(w http.ResponseWriter, r *http.Request, err error) {
	status, detail := http.StatusInternalServerError, "internal error"
	var httpErr *httpx.HTTPError
	var fields *validator.ErrorValidation
	switch {
	case errors.As(err, &httpErr):
		status, detail = httpErr.Status, httpErr.Message
	case errors.As(err, &fields):
		status, detail = http.StatusBadRequest, fields.Error()
	case errors.Is(err, user.ErrEmailTaken):
		status, detail = http.StatusConflict, user.ErrEmailTaken.Error()
	case errors.Is(err, ErrInvalidCredentials):
		status, detail = http.StatusUnauthorized, ErrInvalidCredentials.Error()
	case errors.Is(err, ErrUnauthenticated):
		status, detail = http.StatusUnauthorized, ErrUnauthenticated.Error()
	}
	if status == http.StatusInternalServerError {
		h.log.ErrorContext(r.Context(), "auth request failed", "err", err, "method", r.Method, "path", r.URL.Path)
	}
	if err := httpx.SendError(w, httpx.NewErrorResponse(http.StatusText(status), status, detail, r.URL.Path, fields)); err != nil {
		h.log.ErrorContext(r.Context(), "write auth error response", "err", err)
	}
}
