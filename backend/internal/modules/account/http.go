package account

import (
	"backend/internal/platform/transport/httpx"
	"backend/internal/platform/transport/httpx/middleware"
	"log/slog"
	"net/http"
)

type Handler struct {
	accountService *Service
	log            *slog.Logger
}

func NewHandler(accountService *Service, log *slog.Logger) *Handler {
	return &Handler{accountService: accountService, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /accounts", middleware.AuthGuard(httpx.Adapt(h.createAccount, h.log)))
	mux.Handle("GET /accounts", middleware.AuthGuard(httpx.Adapt(h.getAccounts, h.log)))
	mux.Handle("GET /accounts/{id}", middleware.AuthGuard(httpx.Adapt(h.getAccount, h.log)))
}

func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request) error {
	httpx.NoStoreHeader(w)
	var payload CreateAccountDTO
	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		return httpx.DecodeError(err)
	}
	if errs := payload.Validate(); len(errs) > 0 {
		return httpx.ValidationError(errs)
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	data, err := h.accountService.CreateUserAccount(r.Context(), userID, &payload)
	if err != nil {
		return err
	}
	httpx.WriteJSON(w, http.StatusCreated, data)
	return nil
}

func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) error {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())

	data, err := h.accountService.GetUserAccount(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, data)
	return nil
}

func (h *Handler) getAccounts(w http.ResponseWriter, r *http.Request) error {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())

	data, err := h.accountService.GetUserAccounts(r.Context(), userID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, data)
	return nil
}
