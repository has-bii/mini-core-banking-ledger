package wallet

import (
	"backend/internal/platform/httpx"
	"backend/internal/platform/middleware"
	"backend/internal/platform/validator"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	walletService *Service
	log           *slog.Logger
}

func NewHandler(walletService *Service, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{walletService: walletService, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /wallets", middleware.AuthGuard(h.createWallet))
	mux.Handle("GET /wallets", middleware.AuthGuard(h.getWallets))
	mux.Handle("GET /wallets/{id}", middleware.AuthGuard(h.getWallet))
}

func (h *Handler) createWallet(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	var payload CreateWalletDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		h.sendError(w, r, err)
		return
	}
	if err := payload.Validate(); err != nil {
		h.sendError(w, r, err)
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	data, err := h.walletService.CreateUserWallet(r.Context(), userID, &payload)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusCreated, data)
}

func (h *Handler) getWallet(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())
	data, err := h.walletService.GetUserWallet(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusOK, data)
}

func (h *Handler) getWallets(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())
	data, err := h.walletService.GetUserWallets(r.Context(), userID)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusOK, data)
}

func (h *Handler) sendSuccess(w http.ResponseWriter, r *http.Request, status int, data any) {
	if err := httpx.SendSuccess(w, status, httpx.NewSuccessResponse("", data)); err != nil {
		h.log.ErrorContext(r.Context(), "write wallet response", "err", err)
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
	case errors.Is(err, ErrInvalidWalletID):
		status, detail = http.StatusBadRequest, ErrInvalidWalletID.Error()
	case errors.Is(err, ErrInvalidUserID):
		status, detail = http.StatusBadRequest, ErrInvalidUserID.Error()
	}
	if status == http.StatusInternalServerError {
		h.log.ErrorContext(r.Context(), "wallet request failed", "err", err, "method", r.Method, "path", r.URL.Path)
	}
	if err := httpx.SendError(w, httpx.NewErrorResponse(http.StatusText(status), status, detail, r.URL.Path, fields)); err != nil {
		h.log.ErrorContext(r.Context(), "write wallet error response", "err", err)
	}
}
