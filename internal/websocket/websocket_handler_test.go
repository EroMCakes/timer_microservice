package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"timer-microservice/internal/types"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTimerService is a mock of TimerService
type MockTimerService struct {
	mock.Mock
}

func (m *MockTimerService) CreateTimer(sessionID string, maxTime int64) (*types.Timer, error) {
	args := m.Called(sessionID, maxTime)
	return args.Get(0).(*types.Timer), args.Error(1)
}

func (m *MockTimerService) PauseTimer(id uint) (*types.Timer, error) {
	args := m.Called(id)
	return args.Get(0).(*types.Timer), args.Error(1)
}

func (m *MockTimerService) ResumeTimer(id uint) (*types.Timer, error) {
	args := m.Called(id)
	return args.Get(0).(*types.Timer), args.Error(1)
}

func (m *MockTimerService) StopTimer(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTimerService) ModifyTimer(id uint, newMaxTime int64) (*types.Timer, error) {
	args := m.Called(id, newMaxTime)
	return args.Get(0).(*types.Timer), args.Error(1)
}

func setupWebSocketServer(t *testing.T) (*httptest.Server, *Handler, *MockTimerService) {
	mockService := new(MockTimerService)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	handler := NewHandler(mockService, sugar)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.HandleGameMasterWebSocket(w, r)
	}))

	return server, handler, mockService
}

func connectWebSocket(t *testing.T, server *httptest.Server) *websocket.Conn {
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	return ws
}

func TestCreateTimer(t *testing.T) {
	server, _, mockService := setupWebSocketServer(t)
	defer server.Close()

	ws := connectWebSocket(t, server)
	defer ws.Close()

	createMsg := types.WebSocketMessage{
		Type: types.TypeTimerCreate,
		Payload: json.RawMessage(`{
			"sessionId": "test-session",
			"maxTime": 60
		}`),
	}

	expectedTimer := &types.Timer{
		ID:          1,
		SessionID:   "test-session",
		MaxTime:     60,
		CurrentTime: 60,
		IsPaused:    false,
	}
	mockService.On("CreateTimer", "test-session", int64(60)).Return(expectedTimer, nil)

	err := ws.WriteJSON(createMsg)
	assert.NoError(t, err)

	var response types.WebSocketMessage
	err = ws.ReadJSON(&response)
	assert.NoError(t, err)
	assert.Equal(t, types.TypeTimerUpdate, response.Type)

	var timerResponse types.Timer
	err = json.Unmarshal(response.Payload, &timerResponse)
	assert.NoError(t, err)
	assert.Equal(t, expectedTimer.ID, timerResponse.ID)
	assert.Equal(t, expectedTimer.SessionID, timerResponse.SessionID)
	assert.Equal(t, expectedTimer.MaxTime, timerResponse.MaxTime)
	assert.Equal(t, expectedTimer.CurrentTime, timerResponse.CurrentTime)
	assert.Equal(t, expectedTimer.IsPaused, timerResponse.IsPaused)

	mockService.AssertExpectations(t)
}

func TestPauseTimer(t *testing.T) {
	server, _, mockService := setupWebSocketServer(t)
	defer server.Close()

	ws := connectWebSocket(t, server)
	defer ws.Close()

	pauseMsg := types.WebSocketMessage{
		Type: types.TypeTimerPause,
		Payload: json.RawMessage(`{
			"sessionId": "1"
		}`),
	}

	pausedTimer := &types.Timer{
		ID:          1,
		SessionID:   "test-session",
		MaxTime:     60,
		CurrentTime: 30,
		IsPaused:    true,
	}
	mockService.On("PauseTimer", uint(1)).Return(pausedTimer, nil)

	err := ws.WriteJSON(pauseMsg)
	assert.NoError(t, err)

	var response types.WebSocketMessage
	err = ws.ReadJSON(&response)
	assert.NoError(t, err)
	assert.Equal(t, types.TypeTimerUpdate, response.Type)

	var timerResponse types.Timer
	err = json.Unmarshal(response.Payload, &timerResponse)
	assert.NoError(t, err)
	assert.Equal(t, pausedTimer.ID, timerResponse.ID)
	assert.Equal(t, pausedTimer.IsPaused, timerResponse.IsPaused)

	mockService.AssertExpectations(t)
}

func TestResumeTimer(t *testing.T) {
	server, _, mockService := setupWebSocketServer(t)
	defer server.Close()

	ws := connectWebSocket(t, server)
	defer ws.Close()

	resumeMsg := types.WebSocketMessage{
		Type: types.TypeTimerResume,
		Payload: json.RawMessage(`{
			"sessionId": "1"
		}`),
	}

	resumedTimer := &types.Timer{
		ID:          1,
		SessionID:   "test-session",
		MaxTime:     60,
		CurrentTime: 30,
		IsPaused:    false,
	}
	mockService.On("ResumeTimer", uint(1)).Return(resumedTimer, nil)

	err := ws.WriteJSON(resumeMsg)
	assert.NoError(t, err)

	var response types.WebSocketMessage
	err = ws.ReadJSON(&response)
	assert.NoError(t, err)
	assert.Equal(t, types.TypeTimerUpdate, response.Type)

	var timerResponse types.Timer
	err = json.Unmarshal(response.Payload, &timerResponse)
	assert.NoError(t, err)
	assert.Equal(t, resumedTimer.ID, timerResponse.ID)
	assert.Equal(t, resumedTimer.IsPaused, timerResponse.IsPaused)

	mockService.AssertExpectations(t)
}

func TestModifyTimer(t *testing.T) {
	server, _, mockService := setupWebSocketServer(t)
	defer server.Close()

	ws := connectWebSocket(t, server)
	defer ws.Close()

	modifyMsg := types.WebSocketMessage{
		Type: types.TypeTimerModify,
		Payload: json.RawMessage(`{
			"sessionId": "1",
			"maxTime": 90
		}`),
	}

	modifiedTimer := &types.Timer{
		ID:          1,
		SessionID:   "test-session",
		MaxTime:     90,
		CurrentTime: 60,
		IsPaused:    false,
	}
	mockService.On("ModifyTimer", uint(1), int64(90)).Return(modifiedTimer, nil)

	err := ws.WriteJSON(modifyMsg)
	assert.NoError(t, err)

	var response types.WebSocketMessage
	err = ws.ReadJSON(&response)
	assert.NoError(t, err)
	assert.Equal(t, types.TypeTimerUpdate, response.Type)

	var timerResponse types.Timer
	err = json.Unmarshal(response.Payload, &timerResponse)
	assert.NoError(t, err)
	assert.Equal(t, modifiedTimer.ID, timerResponse.ID)
	assert.Equal(t, modifiedTimer.MaxTime, timerResponse.MaxTime)

	mockService.AssertExpectations(t)
}

func TestStopTimer(t *testing.T) {
	server, _, mockService := setupWebSocketServer(t)
	defer server.Close()

	ws := connectWebSocket(t, server)
	defer ws.Close()

	stopMsg := types.WebSocketMessage{
		Type: types.TypeTimerStop,
		Payload: json.RawMessage(`{
			"sessionId": "1"
		}`),
	}

	mockService.On("StopTimer", uint(1)).Return(nil)

	err := ws.WriteJSON(stopMsg)
	assert.NoError(t, err)

	var response types.WebSocketMessage
	err = ws.ReadJSON(&response)
	assert.NoError(t, err)
	assert.Equal(t, types.TypeTimerStop, response.Type)

	var stopResponse struct {
		ID uint `json:"id"`
	}
	err = json.Unmarshal(response.Payload, &stopResponse)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), stopResponse.ID)

	mockService.AssertExpectations(t)
}

func TestBroadcastTimerUpdate(t *testing.T) {
	server, handler, _ := setupWebSocketServer(t)
	defer server.Close()

	ws1 := connectWebSocket(t, server)
	defer ws1.Close()

	ws2 := connectWebSocket(t, server)
	defer ws2.Close()

	updateTimer := &types.Timer{
		ID:          1,
		SessionID:   "test-session",
		MaxTime:     60,
		CurrentTime: 30,
		IsPaused:    false,
	}

	// Use a goroutine to broadcast the update
	go handler.BroadcastTimerUpdate(updateTimer)

	// Function to read and verify the response
	verifyResponse := func(ws *websocket.Conn) {
		var response types.WebSocketMessage
		err := ws.ReadJSON(&response)
		assert.NoError(t, err)
		assert.Equal(t, types.TypeTimerUpdate, response.Type)

		var timerResponse types.Timer
		err = json.Unmarshal(response.Payload, &timerResponse)
		assert.NoError(t, err)
		assert.Equal(t, updateTimer.ID, timerResponse.ID)
		assert.Equal(t, updateTimer.CurrentTime, timerResponse.CurrentTime)
	}

	// Use channels to synchronize goroutines
	done1 := make(chan bool)
	done2 := make(chan bool)

	go func() {
		verifyResponse(ws1)
		done1 <- true
	}()

	go func() {
		verifyResponse(ws2)
		done2 <- true
	}()

	// Wait for both goroutines to complete or timeout
	select {
	case <-done1:
		<-done2
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out")
	}
}
