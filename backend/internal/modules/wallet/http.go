package wallet

import (
	"backend/internal/platform/httpx"
	"backend/internal/platform/middleware"
	"context"
	"errors"
	"net/http"
)

type walletService interface {
	CreateUserWallet(ctx context.Context, userID string, data CreateWalletDTO) (Wallet, error)
	GetUserWallet(ctx context.Context, walletID, userID string) (Wallet, error)
	GetUserWallets(ctx context.Context, userID string) ([]Wallet, error)
}

type Handler struct {
	svc walletService
}

func NewHandler(svc walletService) *Handler {
	return &Handler{svc: svc}
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
	data, err := h.svc.CreateUserWallet(ctx, userID, payload)
	if err != nil {
		h.sendError(w, r, err)
		return
	}
	httpx.SendSuccess(w, r, http.StatusCreated, httpx.SuccessResponse{Data: data})
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
	httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{Data: data})
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
	httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{Data: data})
}

func (h *Handler) sendError(w http.ResponseWriter, r *http.Request, err error) {
	httpx.SendRequestError(w, r, "wallet", err, domainProblem)
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
