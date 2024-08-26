package websocket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"timer-microservice/internal/types"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

type TimerServiceInterface interface {
	CreateTimer(sessionID string, maxTime int64) (*types.Timer, error)
	PauseTimer(id uint) (*types.Timer, error)
	ResumeTimer(id uint) (*types.Timer, error)
	StopTimer(id uint) error
	ModifyTimer(id uint, newMaxTime int64) (*types.Timer, error)
}

type Handler struct {
	service     TimerServiceInterface
	logger      *zap.SugaredLogger
	connections map[string]*websocket.Conn
	mutex       sync.RWMutex
}

func NewHandler(service TimerServiceInterface, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service:     service,
		logger:      logger,
		connections: make(map[string]*websocket.Conn),
	}
}

func (h *Handler) SetService(service TimerServiceInterface) {
	h.service = service
}

func (h *Handler) HandleCustomerWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	h.handleWebSocket(w, r, sessionID, false)
}

func (h *Handler) HandleGameMasterWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	h.handleWebSocket(w, r, sessionID, true)
}

func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request, sessionID string, isGameMaster bool) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorw("Failed to upgrade connection", "error", err)
		return
	}

	h.mutex.Lock()
	h.connections[sessionID] = conn
	h.mutex.Unlock()

	h.logger.Infow("New WebSocket connection established", "sessionID", sessionID, "isGameMaster", isGameMaster)

	defer h.closeConnection(sessionID, conn)

	for {
		var wsMessage types.WebSocketMessage
		err := conn.ReadJSON(&wsMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorw("WebSocket read error", "error", err, "sessionID", sessionID)
			}
			break
		}

		h.handleMessage(wsMessage, sessionID, isGameMaster)
	}
}

func (h *Handler) closeConnection(sessionID string, conn *websocket.Conn) {
	h.mutex.Lock()
	delete(h.connections, sessionID)
	h.mutex.Unlock()
	conn.Close()
	h.logger.Infow("WebSocket connection closed", "sessionID", sessionID)
}

func (h *Handler) handleMessage(message types.WebSocketMessage, sessionID string, isGameMaster bool) {
	switch message.Type {
	case types.TypeTimerCreate:
		h.handleTimerCreate(message.Payload, isGameMaster)
	case types.TypeTimerPause:
		h.handleTimerPause(message.Payload, isGameMaster)
	case types.TypeTimerResume:
		h.handleTimerResume(message.Payload, isGameMaster)
	case types.TypeTimerStop:
		h.handleTimerStop(message.Payload, isGameMaster)
	case types.TypeTimerModify:
		h.handleTimerModify(message.Payload, isGameMaster)
	default:
		h.logger.Warnw("Unknown message type received", "type", message.Type, "sessionID", sessionID)
	}
}

func (h *Handler) handleTimerCreate(payload json.RawMessage, isGameMaster bool) {
	var createPayload types.TimerRequest
	if err := json.Unmarshal(payload, &createPayload); err != nil {
		h.logger.Errorw("Failed to unmarshal timer create payload", "error", err)
		return
	}
	timer, err := h.service.CreateTimer(createPayload.SessionID, createPayload.MaxTime)
	if err != nil {
		h.logger.Errorw("Failed to create timer", "error", err)
		return
	}
	h.broadcastTimerUpdate(timer, isGameMaster)
}

// Implement similar handler functions for pause, resume, stop, and modify

func (h *Handler) handleTimerPause(payload json.RawMessage, isGameMaster bool) {
	var pausePayload types.TimerRequest
	if err := json.Unmarshal(payload, &pausePayload); err != nil {
		h.logger.Errorw("Failed to unmarshal timer pause payload", "error", err)
		return
	}

	id, err := strconv.ParseUint(pausePayload.SessionID, 10, 64)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		return
	}

	timer, err := h.service.PauseTimer(uint(id))
	if err != nil {
		h.logger.Errorw("Failed to pause timer", "error", err)
		return
	}

	h.broadcastTimerUpdate(timer, isGameMaster)
}

func (h *Handler) handleTimerResume(payload json.RawMessage, isGameMaster bool) {
	var resumePayload types.TimerRequest
	if err := json.Unmarshal(payload, &resumePayload); err != nil {
		h.logger.Errorw("Failed to unmarshal timer resume payload", "error", err)
		return
	}

	id, err := strconv.ParseUint(resumePayload.SessionID, 10, 64)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		return
	}

	timer, err := h.service.ResumeTimer(uint(id))
	if err != nil {
		h.logger.Errorw("Failed to resume timer", "error", err)
		return
	}

	h.broadcastTimerUpdate(timer, isGameMaster)
}

func (h *Handler) handleTimerStop(payload json.RawMessage, isGameMaster bool) {
	var stopPayload types.TimerRequest
	if err := json.Unmarshal(payload, &stopPayload); err != nil {
		h.logger.Errorw("Failed to unmarshal timer stop payload", "error", err)
		return
	}

	id, err := strconv.ParseUint(stopPayload.SessionID, 10, 64)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		return
	}

	err = h.service.StopTimer(uint(id))
	if err != nil {
		h.logger.Errorw("Failed to stop timer", "error", err)
		return
	}

	h.broadcastTimerStop(uint(id), isGameMaster)
}

func (h *Handler) handleTimerModify(payload json.RawMessage, isGameMaster bool) {
	var modifyPayload types.TimerRequest
	if err := json.Unmarshal(payload, &modifyPayload); err != nil {
		h.logger.Errorw("Failed to unmarshal timer modify payload", "error", err)
		return
	}

	id, err := strconv.ParseUint(modifyPayload.SessionID, 10, 64)
	if err != nil {
		h.logger.Errorw("Invalid timer ID", "error", err)
		return
	}

	timer, err := h.service.ModifyTimer(uint(id), modifyPayload.MaxTime)
	if err != nil {
		h.logger.Errorw("Failed to modify timer", "error", err)
		return
	}

	h.broadcastTimerUpdate(timer, isGameMaster)
}

func (h *Handler) broadcastTimerUpdate(timer *types.Timer, isGameMaster bool) {
	payload, err := json.Marshal(timer)
	if err != nil {
		h.logger.Errorw("Failed to marshal timer update payload", "error", err)
		return
	}

	message := types.WebSocketMessage{
		Type:    types.TypeTimerUpdate,
		Payload: json.RawMessage(payload),
	}

	h.broadcastMessage(message, timer.SessionID, isGameMaster)
}

func (h *Handler) broadcastTimerStop(timerID uint, isGameMaster bool) {
	payload, err := json.Marshal(struct {
		ID uint `json:"id"`
	}{ID: timerID})
	if err != nil {
		h.logger.Errorw("Failed to marshal timer stop payload", "error", err)
		return
	}

	message := types.WebSocketMessage{
		Type:    types.TypeTimerStop,
		Payload: json.RawMessage(payload),
	}

	h.broadcastMessage(message, "", isGameMaster)
}

func (h *Handler) broadcastMessage(message types.WebSocketMessage, sessionID string, isGameMaster bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for connSessionID, conn := range h.connections {
		if isGameMaster || connSessionID == sessionID {
			if err := conn.WriteJSON(message); err != nil {
				h.logger.Errorw("Failed to send WebSocket message", "error", err, "sessionID", connSessionID)
				// Consider closing the connection here if it's a persistent error
			}
		}
	}
}

// Add a method to broadcast updates from the timer service
func (h *Handler) BroadcastTimerUpdate(timer *types.Timer) {
	h.broadcastTimerUpdate(timer, true) // Broadcast to all, including game masters
}
