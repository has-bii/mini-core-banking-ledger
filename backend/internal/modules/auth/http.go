package auth

import (
	"backend/internal/modules/user"
	"backend/internal/platform/httpx"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var payload RegisterPayload

	if err := httpx.DecodeJSON(w, r, &payload); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	payload, errs := payload.Validate()
	if len(errs) > 0 {
		httpx.WriteError(w, http.StatusBadRequest, errors.Join(errs...).Error())
		return
	}

	u, err := h.svc.Register(r.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailTaken):
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			h.log.Error("registration failed", "err", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
}
