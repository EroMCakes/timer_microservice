package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"timer-microservice/internal/repository"
	"timer-microservice/internal/types"
	"timer-microservice/internal/websocket"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type TimerService struct {
	repo      repository.TimerRepository
	logger    *zap.SugaredLogger
	redis     *redis.Client
	stopChan  chan struct{}
	wsHandler websocket.HandlerInterface
}

type TimerServiceInterface interface {
	StartTimerUpdates()
	StopTimerUpdates()
	CreateTimer(sessionID string, maxTime int64) (*types.Timer, error)
	PauseTimer(id uint) (*types.Timer, error)
	ResumeTimer(id uint) (*types.Timer, error)
	StopTimer(id uint) error
	ModifyTimer(id uint, newMaxTime int64) (*types.Timer, error)
	GetAllTimers() ([]types.Timer, error)
	RestoreTimers() error
}

func NewTimerService(repo repository.TimerRepository, logger *zap.SugaredLogger, redisClient *redis.Client, wsHandler websocket.HandlerInterface) TimerServiceInterface {
	return &TimerService{
		repo:      repo,
		logger:    logger,
		redis:     redisClient,
		stopChan:  make(chan struct{}),
		wsHandler: wsHandler,
	}
}

func (s *TimerService) StartTimerUpdates() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateTimers()
		case <-s.stopChan:
			return
		}
	}
}

func (s *TimerService) updateTimers() {
	timers, err := s.repo.GetActiveTimers()
	if err != nil {
		s.logger.Errorw("Failed to get active timers", "error", err)
		return
	}

	for _, timer := range timers {
		if !timer.IsPaused && timer.CurrentTime > 0 {
			timer.CurrentTime--
			if err := s.repo.Update(&timer); err != nil {
				s.logger.Errorw("Failed to update timer", "error", err, "timerID", timer.ID)
				continue
			}
			s.broadcastTimerUpdate(&timer)
		}
	}
}

func (s *TimerService) broadcastTimerUpdate(timer *types.Timer) {
	// Log the update
	s.logger.Infow("Timer updated", "timerID", timer.ID, "currentTime", timer.CurrentTime)

	// Convert the models.Timer to types.Timer if necessary
	// This step might be needed if your WebSocket handler expects a different type
	wsTimer := &types.Timer{
		ID:          timer.ID,
		SessionID:   timer.SessionID,
		CurrentTime: timer.CurrentTime,
		MaxTime:     timer.MaxTime,
		IsPaused:    timer.IsPaused,
	}

	// Use the WebSocket handler to broadcast the update
	s.wsHandler.BroadcastTimerUpdate(wsTimer)
}

func (s *TimerService) StopTimerUpdates() {
	close(s.stopChan)
}

func (s *TimerService) CreateTimer(sessionID string, maxTime int64) (*types.Timer, error) {
	timer := &types.Timer{
		SessionID:   sessionID,
		MaxTime:     maxTime,
		CurrentTime: maxTime,
		IsPaused:    false,
	}

	err := s.repo.Create(timer)
	if err != nil {
		s.logger.Errorw("Failed to create timer", "error", err, "sessionID", sessionID)
		return nil, err
	}

	s.persistTimer(timer)

	return timer, nil
}

func (s *TimerService) PauseTimer(id uint) (*types.Timer, error) {
	timer, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Errorw("Failed to find timer", "error", err, "id", id)
		return nil, err
	}

	timer.IsPaused = true
	err = s.repo.Update(timer)
	if err != nil {
		s.logger.Errorw("Failed to pause timer", "error", err, "id", id)
		return nil, err
	}

	s.persistTimer(timer)

	return timer, nil
}

func (s *TimerService) ResumeTimer(id uint) (*types.Timer, error) {
	timer, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Errorw("Failed to find timer", "error", err, "id", id)
		return nil, err
	}

	timer.IsPaused = false
	err = s.repo.Update(timer)
	if err != nil {
		s.logger.Errorw("Failed to resume timer", "error", err, "id", id)
		return nil, err
	}

	s.persistTimer(timer)

	return timer, nil
}

func (s *TimerService) StopTimer(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Errorw("Failed to stop timer", "error", err, "id", id)
		return err
	}

	s.redis.Del(context.Background(), "timer:"+fmt.Sprint(id))

	return nil
}

func (s *TimerService) ModifyTimer(id uint, newMaxTime int64) (*types.Timer, error) {
	timer, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Errorw("Failed to find timer", "error", err, "id", id)
		return nil, err
	}

	timer.MaxTime = newMaxTime
	timer.CurrentTime = newMaxTime
	err = s.repo.Update(timer)
	if err != nil {
		s.logger.Errorw("Failed to modify timer", "error", err, "id", id)
		return nil, err
	}

	s.persistTimer(timer)

	return timer, nil
}

func (s *TimerService) GetAllTimers() ([]types.Timer, error) {
	return s.repo.FindAll()
}

func (s *TimerService) persistTimer(timer *types.Timer) {
	timerJSON, err := json.Marshal(timer)
	if err != nil {
		s.logger.Errorw("Failed to marshal timer", "error", err)
		return
	}

	err = s.redis.Set(context.Background(), "timer:"+timer.SessionID, timerJSON, 24*time.Hour).Err()
	if err != nil {
		s.logger.Errorw("Failed to persist timer to Redis", "error", err)
	}
}

func (s *TimerService) RestoreTimers() error {
	keys, err := s.redis.Keys(context.Background(), "timer:*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		timerJSON, err := s.redis.Get(context.Background(), key).Result()
		if err != nil {
			s.logger.Errorw("Failed to get timer from Redis", "error", err)
			continue
		}

		var timer types.Timer
		err = json.Unmarshal([]byte(timerJSON), &timer)
		if err != nil {
			s.logger.Errorw("Failed to unmarshal timer", "error", err)
			continue
		}

		err = s.repo.Update(&timer)
		if err != nil {
			s.logger.Errorw("Failed to restore timer", "error", err)
			continue
		}
	}

	return nil
}
