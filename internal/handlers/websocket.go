package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"timer-microservice/internal/service"
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

type WebSocketHandler struct {
	service     service.TimerService
	logger      *zap.SugaredLogger
	connections map[string]*websocket.Conn
	mutex       sync.RWMutex
}

func NewWebSocketHandler(service service.TimerService, logger *zap.SugaredLogger) *WebSocketHandler {
	return &WebSocketHandler{
		service:     service,
		logger:      logger,
		connections: make(map[string]*websocket.Conn),
	}
}

func (h *WebSocketHandler) HandleCustomerWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorw("Failed to upgrade connection", "error", err)
		return
	}

	h.handleConnection(conn, sessionID, false)
}

func (h *WebSocketHandler) HandleGameMasterWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorw("Failed to upgrade connection", "error", err)
		return
	}

	h.handleConnection(conn, sessionID, true)
}

func (h *WebSocketHandler) handleConnection(conn *websocket.Conn, sessionID string, isGameMaster bool) {
	h.mutex.Lock()
	h.connections[sessionID] = conn
	h.mutex.Unlock()

	defer func() {
		h.mutex.Lock()
		delete(h.connections, sessionID)
		h.mutex.Unlock()
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.logger.Errorw("WebSocket read error", "error", err)
			break
		}

		var wsMessage types.WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			h.logger.Errorw("Failed to unmarshal WebSocket message", "error", err)
			continue
		}

		h.handleMessage(wsMessage, sessionID, isGameMaster)
	}
}

func (h *WebSocketHandler) handleMessage(message types.WebSocketMessage, sessionID string, isGameMaster bool) {
	switch message.Type {
	case types.TypeTimerCreate:
		var createPayload types.TimerRequest
		json.Unmarshal(message.Payload, &createPayload)
		timer, err := h.service.CreateTimer(createPayload.SessionID, createPayload.MaxTime)
		if err != nil {
			h.logger.Errorw("Failed to create timer", "error", err)
			return
		}
		h.broadcastTimerUpdate(timer, isGameMaster)

	case types.TypeTimerPause:
		var pausePayload struct{ ID uint }
		json.Unmarshal(message.Payload, &pausePayload)
		timer, err := h.service.PauseTimer(pausePayload.ID)
		if err != nil {
			h.logger.Errorw("Failed to pause timer", "error", err)
			return
		}
		h.broadcastTimerUpdate(timer, isGameMaster)

	case types.TypeTimerResume:
		var resumePayload struct{ ID uint }
		json.Unmarshal(message.Payload, &resumePayload)
		timer, err := h.service.ResumeTimer(resumePayload.ID)
		if err != nil {
			h.logger.Errorw("Failed to resume timer", "error", err)
			return
		}
		h.broadcastTimerUpdate(timer, isGameMaster)

	case types.TypeTimerStop:
		var stopPayload struct{ ID uint }
		json.Unmarshal(message.Payload, &stopPayload)
		err := h.service.StopTimer(stopPayload.ID)
		if err != nil {
			h.logger.Errorw("Failed to stop timer", "error", err)
			return
		}
		h.broadcastTimerStop(stopPayload.ID, isGameMaster)

	case types.TypeTimerModify:
		var modifyPayload struct {
			ID         uint  `json:"id"`
			NewMaxTime int64 `json:"newMaxTime"`
		}
		json.Unmarshal(message.Payload, &modifyPayload)
		timer, err := h.service.ModifyTimer(modifyPayload.ID, modifyPayload.NewMaxTime)
		if err != nil {
			h.logger.Errorw("Failed to modify timer", "error", err)
			return
		}
		h.broadcastTimerUpdate(timer, isGameMaster)
	}
}

func (h *WebSocketHandler) broadcastTimerUpdate(timer *types.Timer, isGameMaster bool) {
	payload, err := json.Marshal(timer)
	if err != nil {
		h.logger.Errorw("Failed to marshal timer update payload", "error", err)
		return
	}

	message := types.WebSocketMessage{
		Type:    types.TypeTimerUpdate,
		Payload: json.RawMessage(payload),
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for sessionID, conn := range h.connections {
		if isGameMaster || sessionID == timer.SessionID {
			if err := conn.WriteJSON(message); err != nil {
				h.logger.Errorw("Failed to send WebSocket message", "error", err)
			}
		}
	}
}

func (h *WebSocketHandler) broadcastTimerStop(timerID uint, isGameMaster bool) {
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

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, conn := range h.connections {
		if isGameMaster {
			if err := conn.WriteJSON(message); err != nil {
				h.logger.Errorw("Failed to send WebSocket message", "error", err)
			}
		}
	}
}
