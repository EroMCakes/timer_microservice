package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"timer-microservice/internal/types"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTimerService is a mock of TimerServiceInterface
type MockTimerService struct {
	mock.Mock
}

func (m *MockTimerService) StartTimerUpdates() {
	m.Called()
}

func (m *MockTimerService) StopTimerUpdates() {
	m.Called()
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

func (m *MockTimerService) GetAllTimers() ([]types.Timer, error) {
	args := m.Called()
	return args.Get(0).([]types.Timer), args.Error(1)
}

func (m *MockTimerService) RestoreTimers() error {
	args := m.Called()
	return args.Error(0)
}

func TestCreateTimer(t *testing.T) {
	mockService := new(MockTimerService)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	handler := NewTimerHandler(mockService, sugar)

	req := types.TimerRequest{
		SessionID: "test-session",
		MaxTime:   60,
	}

	expectedTimer := &types.Timer{
		ID:          1,
		SessionID:   "test-session",
		MaxTime:     60,
		CurrentTime: 60,
		IsPaused:    false,
	}

	mockService.On("CreateTimer", req.SessionID, req.MaxTime).Return(expectedTimer, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/timer", bytes.NewBuffer(body))

	handler.CreateTimer(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response types.Timer
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTimer.ID, response.ID)
	assert.Equal(t, expectedTimer.SessionID, response.SessionID)
	assert.Equal(t, expectedTimer.MaxTime, response.MaxTime)
	assert.Equal(t, expectedTimer.CurrentTime, response.CurrentTime)
	assert.Equal(t, expectedTimer.IsPaused, response.IsPaused)

	mockService.AssertExpectations(t)
}

func TestPauseTimer(t *testing.T) {
	mockService := new(MockTimerService)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	handler := NewTimerHandler(mockService, sugar)

	timerID := uint(1)
	expectedTimer := &types.Timer{
		ID:          timerID,
		SessionID:   "test-session",
		MaxTime:     60,
		CurrentTime: 30,
		IsPaused:    true,
	}

	mockService.On("PauseTimer", timerID).Return(expectedTimer, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/timer/1/pause", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.PauseTimer(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Timer
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTimer.ID, response.ID)
	assert.Equal(t, expectedTimer.SessionID, response.SessionID)
	assert.Equal(t, expectedTimer.MaxTime, response.MaxTime)
	assert.Equal(t, expectedTimer.CurrentTime, response.CurrentTime)
	assert.Equal(t, expectedTimer.IsPaused, response.IsPaused)

	mockService.AssertExpectations(t)
}

func TestResumeTimer(t *testing.T) {
	mockService := new(MockTimerService)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	handler := NewTimerHandler(mockService, sugar)

	timerID := uint(1)
	expectedTimer := &types.Timer{
		ID:          timerID,
		SessionID:   "test-session",
		MaxTime:     60,
		CurrentTime: 30,
		IsPaused:    false,
	}

	mockService.On("ResumeTimer", timerID).Return(expectedTimer, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/timer/1/resume", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ResumeTimer(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Timer
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTimer.ID, response.ID)
	assert.Equal(t, expectedTimer.SessionID, response.SessionID)
	assert.Equal(t, expectedTimer.MaxTime, response.MaxTime)
	assert.Equal(t, expectedTimer.CurrentTime, response.CurrentTime)
	assert.Equal(t, expectedTimer.IsPaused, response.IsPaused)

	mockService.AssertExpectations(t)
}

func TestStopTimer(t *testing.T) {
	mockService := new(MockTimerService)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	handler := NewTimerHandler(mockService, sugar)

	timerID := uint(1)

	mockService.On("StopTimer", timerID).Return(nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/timer/1/stop", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.StopTimer(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Timer stopped and deleted", response["message"])

	mockService.AssertExpectations(t)
}

func TestModifyTimer(t *testing.T) {
	mockService := new(MockTimerService)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	handler := NewTimerHandler(mockService, sugar)

	timerID := uint(1)
	newMaxTime := int64(90)
	expectedTimer := &types.Timer{
		ID:          timerID,
		SessionID:   "test-session",
		MaxTime:     newMaxTime,
		CurrentTime: newMaxTime,
		IsPaused:    false,
	}

	mockService.On("ModifyTimer", timerID, newMaxTime).Return(expectedTimer, nil)

	req := types.TimerRequest{
		MaxTime: newMaxTime,
	}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/timer/1/modify", bytes.NewBuffer(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ModifyTimer(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Timer
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTimer.ID, response.ID)
	assert.Equal(t, expectedTimer.SessionID, response.SessionID)
	assert.Equal(t, expectedTimer.MaxTime, response.MaxTime)
	assert.Equal(t, expectedTimer.CurrentTime, response.CurrentTime)
	assert.Equal(t, expectedTimer.IsPaused, response.IsPaused)

	mockService.AssertExpectations(t)
}
