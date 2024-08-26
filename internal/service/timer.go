package service

import (
	"context"
	"encoding/json"
	"time"

	"timer-microservice/internal/repository"
	"timer-microservice/internal/types"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type TimerService interface {
	CreateTimer(sessionID string, maxTime int64) (*types.Timer, error)
	PauseTimer(id uint) (*types.Timer, error)
	ResumeTimer(id uint) (*types.Timer, error)
	StopTimer(id uint) error
	ModifyTimer(id uint, newMaxTime int64) (*types.Timer, error)
	GetAllTimers() ([]types.Timer, error)
	RestoreTimers() error
}

type timerService struct {
	repo   repository.TimerRepository
	logger *zap.SugaredLogger
	redis  *redis.Client
}

func NewTimerService(repo repository.TimerRepository, logger *zap.SugaredLogger, redisClient *redis.Client) TimerService {
	return &timerService{repo: repo, logger: logger, redis: redisClient}
}

func (s *timerService) CreateTimer(sessionID string, maxTime int64) (*types.Timer, error) {
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

func (s *timerService) PauseTimer(id uint) (*types.Timer, error) {
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

func (s *timerService) ResumeTimer(id uint) (*types.Timer, error) {
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

func (s *timerService) StopTimer(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Errorw("Failed to stop timer", "error", err, "id", id)
		return err
	}

	s.redis.Del(context.Background(), "timer:"+string(id))

	return nil
}

func (s *timerService) ModifyTimer(id uint, newMaxTime int64) (*types.Timer, error) {
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

func (s *timerService) GetAllTimers() ([]types.Timer, error) {
	return s.repo.FindAll()
}

func (s *timerService) persistTimer(timer *types.Timer) {
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

func (s *timerService) RestoreTimers() error {
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
