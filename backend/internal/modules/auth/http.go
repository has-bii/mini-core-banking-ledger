package auth

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/transport/httpx"
	"backend/internal/platform/transport/httpx/middleware"
	"context"
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
	mux.HandleFunc("POST /auth/register", httpx.Adapt(h.register, h.log))
	mux.HandleFunc("POST /auth/login", httpx.Adapt(h.login, h.log))
	mux.HandleFunc("POST /auth/logout", httpx.Adapt(h.logout, h.log))
	mux.Handle("GET /auth/me", middleware.AuthGuard(httpx.Adapt(h.me, h.log)))
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) error {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())
	u, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		return err
	}
	httpx.WriteJSON(w, http.StatusOK, u)
	return nil
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) error {
	httpx.NoStoreHeader(w)
	if cookie, err := r.Cookie("session"); err == nil {
		if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
			return err
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: "", Path: "/",
		HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
		Expires: time.Unix(1, 0).UTC(), MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) error {
	httpx.NoStoreHeader(w)
	var payload LoginDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		return httpx.DecodeError(err)
	}
	if errs := payload.Validate(); len(errs) > 0 {
		return httpx.ValidationError(errs)
	}

	u, created, token, err := h.svc.Login(r.Context(), payload)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: token, Path: "/",
		HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
		Expires: created.ExpiresAt, MaxAge: int(session.Lifetime.Seconds()),
	})
	httpx.WriteJSON(w, http.StatusOK, u)
	return nil
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) error {
	var payload RegisterDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		return httpx.DecodeError(err)
	}
	if errs := payload.Validate(); len(errs) > 0 {
		return httpx.ValidationError(errs)
	}

	u, err := h.svc.Register(r.Context(), payload)
	if err != nil {
		return err
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
	return nil
}
