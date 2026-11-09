package auth

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/httpx"
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type authService interface {
	Register(context.Context, RegisterDTO) (user.User, error)
	Login(context.Context, LoginDTO) (user.User, session.Session, string, error)
	Me(context.Context, string) (user.User, error)
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
	mux.HandleFunc("GET /auth/me", h.me)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", ErrUnauthenticated.Error())
		return
	}

	u, err := h.svc.Me(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, ErrUnauthenticated) {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", ErrUnauthenticated.Error())
		} else {
			h.log.Error("current user lookup failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var payload LoginDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}
	if errs := payload.Validate(); len(errs) > 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", errors.Join(errs...).Error())
		return
	}

	u, created, token, err := h.svc.Login(r.Context(), payload)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", ErrInvalidCredentials.Error())
		} else {
			h.log.Error("login failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
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
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}
	if errs := payload.Validate(); len(errs) > 0 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", errors.Join(errs...).Error())
		return
	}

	u, err := h.svc.Register(r.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailTaken):
			httpx.WriteError(w, http.StatusConflict, "EMAIL_TAKEN", err.Error())
		default:
			h.log.Error("registration failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
}
