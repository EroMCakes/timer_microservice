package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"timer-microservice/internal/service"
	"timer-microservice/internal/types"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type TimerHandler struct {
	service service.TimerService
	logger  *zap.SugaredLogger
}

func NewTimerHandler(service service.TimerService, logger *zap.SugaredLogger) *TimerHandler {
	return &TimerHandler{service: service, logger: logger}
}

func (h *TimerHandler) CreateTimer(w http.ResponseWriter, r *http.Request) {
	var req types.TimerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorw("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timer, err := h.service.CreateTimer(req.SessionID, req.MaxTime)
	if err != nil {
		h.logger.Errorw("Failed to create timer", "error", err)
		http.Error(w, "Failed to create timer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(timer)
}

func (h *TimerHandler) PauseTimer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		http.Error(w, "Invalid timer ID", http.StatusBadRequest)
		return
	}

	timer, err := h.service.PauseTimer(uint(id))
	if err != nil {
		h.logger.Errorw("Failed to pause timer", "error", err, "id", id)
		http.Error(w, "Failed to pause timer", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(timer)
}

func (h *TimerHandler) ResumeTimer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		http.Error(w, "Invalid timer ID", http.StatusBadRequest)
		return
	}

	timer, err := h.service.ResumeTimer(uint(id))
	if err != nil {
		h.logger.Errorw("Failed to resume timer", "error", err, "id", id)
		http.Error(w, "Failed to resume timer", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(timer)
}

func (h *TimerHandler) StopTimer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		http.Error(w, "Invalid timer ID", http.StatusBadRequest)
		return
	}

	err = h.service.StopTimer(uint(id))
	if err != nil {
		h.logger.Errorw("Failed to stop timer", "error", err, "id", id)
		http.Error(w, "Failed to stop timer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Timer stopped and deleted"})
}

func (h *TimerHandler) ModifyTimer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		http.Error(w, "Invalid timer ID", http.StatusBadRequest)
		return
	}

	var req types.TimerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorw("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timer, err := h.service.ModifyTimer(uint(id), req.MaxTime)
	if err != nil {
		h.logger.Errorw("Failed to modify timer", "error", err, "id", id)
		http.Error(w, "Failed to modify timer", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(timer)
}
