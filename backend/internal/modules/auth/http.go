package auth

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/httpx"
	"backend/internal/platform/middleware"

	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type authService interface {
	Register(context.Context, RegisterDTO) (user.User, error)
	Login(context.Context, LoginDTO) (user.User, session.Session, string, error)
	Me(context.Context, string) (user.User, error)
	Logout(context.Context, string) error
}

type Handler struct {
	svc             authService
	secureCookie    bool
	sessionLifetime time.Duration
}

// NewHandler returns an auth handler configured with the given service and cookie settings.
// It requires a positive, whole-second session lifetime for login cookies.
func NewHandler(svc authService, secureCookie bool, sessionLifetime time.Duration) *Handler {
	return &Handler{svc: svc, secureCookie: secureCookie, sessionLifetime: sessionLifetime}
}

// Register adds the auth routes to mux and protects GET /auth/me with the auth guard.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("POST /auth/logout", h.logout)
	mux.Handle("GET /auth/me", middleware.AuthGuard(h.me))
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	ctx := r.Context()
	userID, _ := middleware.UserIDFromContext(ctx)
	u, err := h.svc.Me(ctx, userID)
	if err != nil {
		httpx.SendError(w, r, mapServiceError(err))
		return
	}
	httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{Data: u})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	if cookie, err := r.Cookie("session"); err == nil {
		if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
			httpx.SendError(w, r, mapServiceError(err))
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
		httpx.SendError(w, r, err)
		return
	}
	if err := payload.Validate(); err != nil {
		httpx.SendError(w, r, err)
		return
	}
	u, created, token, err := h.svc.Login(r.Context(), payload)
	if err != nil {
		httpx.SendError(w, r, mapServiceError(err))
		return
	}
	cookie := http.Cookie{
		Name: "session", Value: token, Path: "/",
		HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
		Expires: created.ExpiresAt,
	}
	// http.Cookie.MaxAge is an int and can overflow for valid lifetimes on 32-bit builds.
	maxAge := strconv.FormatInt(int64(h.sessionLifetime/time.Second), 10)
	w.Header().Add("Set-Cookie", cookie.String()+"; Max-Age="+maxAge)
	httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{Data: u})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var payload RegisterDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		httpx.SendError(w, r, err)
		return
	}
	if err := payload.Validate(); err != nil {
		httpx.SendError(w, r, err)
		return
	}
	u, err := h.svc.Register(r.Context(), payload)
	if err != nil {
		httpx.SendError(w, r, mapServiceError(err))
		return
	}
	httpx.SendSuccess(w, r, http.StatusCreated, httpx.SuccessResponse{Data: u})
}

// mapServiceError attaches public HTTP statuses to known errors and leaves unexpected errors unchanged.
func mapServiceError(err error) error {
	switch {
	case errors.Is(err, user.ErrEmailTaken):
		return httpx.WrapHTTPError(http.StatusConflict, user.ErrEmailTaken.Error(), err)
	case errors.Is(err, ErrInvalidCredentials):
		return httpx.WrapHTTPError(http.StatusUnauthorized, ErrInvalidCredentials.Error(), err)
	case errors.Is(err, ErrUnauthenticated):
		return httpx.WrapHTTPError(http.StatusUnauthorized, ErrUnauthenticated.Error(), err)
	default:
		return err
	}
}
