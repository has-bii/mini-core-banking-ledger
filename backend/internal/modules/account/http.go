package account

import (
	legacyMiddleware "backend/internal/platform/transport/httpx/middleware"
	"backend/internal/platform/v2/httpx"
	"backend/internal/platform/v2/middleware"
	"backend/internal/platform/v2/validator"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	accountService *Service
	log            *slog.Logger
}

func NewHandler(accountService *Service, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{accountService: accountService, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /accounts", h.authenticated(h.createAccount))
	mux.Handle("GET /accounts", h.authenticated(h.getAccounts))
	mux.Handle("GET /accounts/{id}", h.authenticated(h.getAccount))
}

func (h *Handler) authenticated(next func(http.ResponseWriter, *http.Request)) http.Handler {
	guard := middleware.AuthGuard(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// shortcut: bridge legacy sessions for account routes until app middleware migrates to v2.
		if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
			if id, ok := legacyMiddleware.UserIDFromContext(r.Context()); ok && id != "" {
				r = r.WithContext(middleware.WithUserID(r.Context(), id))
			}
		}
		guard.ServeHTTP(w, r)
	})
}

func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	var payload CreateAccountDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		h.sendError(w, r, err)
		return
	}
	if err := payload.Validate(); err != nil {
		h.sendError(w, r, err)
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	data, err := h.accountService.CreateUserAccount(r.Context(), userID, &payload)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusCreated, data)
}

func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())
	data, err := h.accountService.GetUserAccount(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusOK, data)
}

func (h *Handler) getAccounts(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())
	data, err := h.accountService.GetUserAccounts(r.Context(), userID)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusOK, data)
}

func (h *Handler) sendSuccess(w http.ResponseWriter, r *http.Request, status int, data any) {
	if err := httpx.SendSuccess(w, status, httpx.NewSuccessResponse("", data)); err != nil {
		h.log.ErrorContext(r.Context(), "write account response", "err", err)
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
	case errors.Is(err, ErrNotFound):
		status, detail = http.StatusNotFound, ErrNotFound.Error()
	case errors.Is(err, ErrInvalidAccountID):
		status, detail = http.StatusBadRequest, ErrInvalidAccountID.Error()
	case errors.Is(err, ErrInvalidUserID):
		status, detail = http.StatusBadRequest, ErrInvalidUserID.Error()
	}
	if status == http.StatusInternalServerError {
		h.log.ErrorContext(r.Context(), "account request failed", "err", err, "method", r.Method, "path", r.URL.Path)
	}
	if err := httpx.SendError(w, httpx.NewErrorResponse(http.StatusText(status), status, detail, r.URL.Path, fields)); err != nil {
		h.log.ErrorContext(r.Context(), "write account error response", "err", err)
	}
}
