package wallet

import (
	"backend/internal/platform/httpx"
	"backend/internal/platform/middleware"
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type walletService interface {
	CreateUserWallet(ctx context.Context, userID string, data *CreateWalletDTO) (Wallet, error)
	GetUserWallet(ctx context.Context, walletID, userID string) (Wallet, error)
	GetUserWallets(ctx context.Context, userID string) ([]Wallet, error)
}

type Handler struct {
	svc walletService
	log *slog.Logger
}

func NewHandler(svc walletService, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{svc: svc, log: log}
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
	ctx := r.Context()
	userID, _ := middleware.UserIDFromContext(ctx)
	data, err := h.svc.CreateUserWallet(ctx, userID, &payload)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusCreated, data)
}

func (h *Handler) getWallet(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	ctx := r.Context()
	userID, _ := middleware.UserIDFromContext(ctx)
	data, err := h.svc.GetUserWallet(ctx, r.PathValue("id"), userID)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusOK, data)
}

func (h *Handler) getWallets(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	ctx := r.Context()
	userID, _ := middleware.UserIDFromContext(ctx)
	data, err := h.svc.GetUserWallets(ctx, userID)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	h.sendSuccess(w, r, http.StatusOK, data)
}

func (h *Handler) sendSuccess(w http.ResponseWriter, r *http.Request, status int, data any) {
	if err := httpx.SendSuccess(w, status, httpx.SuccessResponse{Data: data}); err != nil {
		h.log.ErrorContext(r.Context(), "write wallet response", "err", err)
	}
}

func (h *Handler) sendError(w http.ResponseWriter, r *http.Request, err error) {
	httpx.SendRequestError(w, r, h.log, "wallet", err, domainProblem)
}

func domainProblem(err error) (int, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, ErrNotFound.Error()
	case errors.Is(err, ErrInvalidWalletID):
		return http.StatusBadRequest, ErrInvalidWalletID.Error()
	case errors.Is(err, ErrInvalidUserID):
		return http.StatusBadRequest, ErrInvalidUserID.Error()
	default:
		return 0, ""
	}
}
