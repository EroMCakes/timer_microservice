package service

import (
	"testing"
	"time"

	"timer-microservice/internal/types"
	"timer-microservice/internal/websocket"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTimerRepository is a mock of TimerRepository
type MockTimerRepository struct {
	mock.Mock
}

func (m *MockTimerRepository) Create(timer *types.Timer) error {
	args := m.Called(timer)
	return args.Error(0)
}

func (m *MockTimerRepository) Update(timer *types.Timer) error {
	args := m.Called(timer)
	return args.Error(0)
}

func (m *MockTimerRepository) FindByID(id uint) (*types.Timer, error) {
	args := m.Called(id)
	return args.Get(0).(*types.Timer), args.Error(1)
}

func (m *MockTimerRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTimerRepository) FindAll() ([]types.Timer, error) {
	args := m.Called()
	return args.Get(0).([]types.Timer), args.Error(1)
}

func (m *MockTimerRepository) GetActiveTimers() ([]types.Timer, error) {
	args := m.Called()
	return args.Get(0).([]types.Timer), args.Error(1)
}

// MockWebSocketHandler is a mock of WebSocket HandlerInterface
type MockWebSocketHandler struct {
	mock.Mock
}

func (m *MockWebSocketHandler) BroadcastTimerUpdate(timer *types.Timer) {
	m.Called(timer)
}

func (m *MockWebSocketHandler) SetService(service websocket.TimerServiceInterface) {
	m.Called(service)
}

func TestCreateTimer(t *testing.T) {
	mockRepo := new(MockTimerRepository)
	mockRedis := redis.NewClient(&redis.Options{})
	mockWS := new(MockWebSocketHandler)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	service := NewTimerService(mockRepo, sugar, mockRedis, mockWS)

	sessionID := "test-session"
	maxTime := int64(60)

	expectedTimer := &types.Timer{
		SessionID:   sessionID,
		MaxTime:     maxTime,
		CurrentTime: maxTime,
		IsPaused:    false,
	}

	mockRepo.On("Create", mock.AnythingOfType("*types.Timer")).Return(nil)
	mockWS.On("SetService", mock.AnythingOfType("*TimerService")).Return()

	timer, err := service.CreateTimer(sessionID, maxTime)

	assert.NoError(t, err)
	assert.Equal(t, expectedTimer.SessionID, timer.SessionID)
	assert.Equal(t, expectedTimer.MaxTime, timer.MaxTime)
	assert.Equal(t, expectedTimer.CurrentTime, timer.CurrentTime)
	assert.Equal(t, expectedTimer.IsPaused, timer.IsPaused)

	mockRepo.AssertExpectations(t)
	mockWS.AssertExpectations(t)
}

// ... (keep the other test functions)

func TestUpdateTimers(t *testing.T) {
	mockRepo := new(MockTimerRepository)
	mockRedis := redis.NewClient(&redis.Options{})
	mockWS := new(MockWebSocketHandler)
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	service := NewTimerService(mockRepo, sugar, mockRedis, mockWS)

	activeTimers := []types.Timer{
		{ID: 1, SessionID: "session1", MaxTime: 60, CurrentTime: 30, IsPaused: false},
		{ID: 2, SessionID: "session2", MaxTime: 120, CurrentTime: 90, IsPaused: false},
	}

	mockRepo.On("GetActiveTimers").Return(activeTimers, nil)
	mockRepo.On("Update", mock.AnythingOfType("*types.Timer")).Return(nil)
	mockWS.On("BroadcastTimerUpdate", mock.AnythingOfType("*types.Timer")).Return()
	mockWS.On("SetService", mock.AnythingOfType("*TimerService")).Return()

	go service.StartTimerUpdates()
	time.Sleep(2 * time.Second) // Allow time for the goroutine to run
	service.StopTimerUpdates()

	mockRepo.AssertExpectations(t)
	mockWS.AssertExpectations(t)
}

// Add more tests as needed
