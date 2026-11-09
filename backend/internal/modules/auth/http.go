package auth

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/httpx"
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
	return &Handler{svc: svc, log: log, secureCookie: secureCookie}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("POST /auth/logout", h.logout)
	mux.Handle("GET /auth/me", httpx.AuthGuard()(http.HandlerFunc(h.me)))
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, _ := httpx.UserIDFromContext(r.Context())
	u, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUnauthenticated) {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthenticated, ErrUnauthenticated.Error())
		} else {
			h.log.Error("current user lookup failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal error")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if cookie, err := r.Cookie("session"); err == nil {
		if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
			h.log.Error("logout failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal error")
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
	w.Header().Set("Cache-Control", "no-store")
	var payload LoginDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidJSON, "Invalid JSON body")
		return
	}
	if errs := payload.Validate(); len(errs) > 0 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidationError, errors.Join(errs...).Error())
		return
	}

	u, created, token, err := h.svc.Login(r.Context(), payload)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeInvalidCredentials, ErrInvalidCredentials.Error())
		} else {
			h.log.Error("login failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal error")
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: token, Path: "/",
		HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
		Expires: created.ExpiresAt, MaxAge: int(session.Lifetime.Seconds()),
	})
	httpx.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var payload RegisterDTO

	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeInvalidJSON, "Invalid JSON body")
		return
	}
	if errs := payload.Validate(); len(errs) > 0 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidationError, errors.Join(errs...).Error())
		return
	}

	u, err := h.svc.Register(r.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailTaken):
			httpx.WriteError(w, http.StatusConflict, httpx.CodeEmailTaken, err.Error())
		default:
			h.log.Error("registration failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal error")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
}
