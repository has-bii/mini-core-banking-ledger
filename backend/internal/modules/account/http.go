package account

import (
	"backend/internal/platform/transport/httpx"
	"backend/internal/platform/transport/httpx/middleware"
	"backend/internal/platform/validation"
	"errors"
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
	mux.Handle("GET /accounts", middleware.AuthGuard()(http.HandlerFunc(h.getAccounts)))
	mux.Handle("GET /accounts/{id}", middleware.AuthGuard()(http.HandlerFunc(h.getAccount)))
}

func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())

	data, err := h.accountService.GetUserAccount(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		switch {
		case errors.Is(err, validation.ErrInvalidUUID):
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid account ID")
		case errors.Is(err, ErrNotFound):
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, ErrNotFound.Error())
		default:
			h.log.Error("get account failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal server error")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, data)
}

func (h *Handler) getAccounts(w http.ResponseWriter, r *http.Request) {
	httpx.NoStoreHeader(w)
	userID, _ := middleware.UserIDFromContext(r.Context())

	data, err := h.accountService.GetUserAccounts(r.Context(), userID)
	if err != nil {
		h.log.Error("get accounts failed", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal server error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, data)
}
