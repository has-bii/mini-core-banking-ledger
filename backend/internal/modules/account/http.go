package account

import (
	"backend/internal/platform/transport/httpx"
	"backend/internal/platform/transport/httpx/middleware"
	"backend/internal/platform/validation"
	"context"
	"log/slog"
	"net/http"
)

type accountService interface {
	GetAccountsByUserID(context.Context, string, validation.QueryOrder) ([]Account, error)
}

type Handler struct {
	accountService accountService
	log            *slog.Logger
}

func NewHandler(accountService accountService, log *slog.Logger) *Handler {
	return &Handler{accountService: accountService, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /accounts", middleware.AuthGuard()(http.HandlerFunc(h.getAccounts)))
}

func (h *Handler) getAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, _ := middleware.UserIDFromContext(r.Context())

	order := validation.ASC
	if r.URL.Query().Has("order") {
		order = validation.QueryOrder(r.URL.Query().Get("order"))
		if err := order.Validate(); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidationError, err.Error())
			return
		}
	}

	data, err := h.accountService.GetAccountsByUserID(r.Context(), userID, order)
	if err != nil {
		h.log.Error("get accounts failed", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal server error")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, data)
}
