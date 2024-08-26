package repository

import (
	"timer-microservice/internal/types"

	"gorm.io/gorm"
)

type TimerRepository interface {
	Create(timer *types.Timer) error
	Update(timer *types.Timer) error
	FindByID(id uint) (*types.Timer, error)
	Delete(id uint) error
	FindAll() ([]types.Timer, error)
}

type timerRepository struct {
	db *gorm.DB
}

func NewTimerRepository(db *gorm.DB) TimerRepository {
	return &timerRepository{db: db}
}

func (r *timerRepository) Create(timer *types.Timer) error {
	return r.db.Create(timer).Error
}

func (r *timerRepository) Update(timer *types.Timer) error {
	return r.db.Save(timer).Error
}

func (r *timerRepository) FindByID(id uint) (*types.Timer, error) {
	var timer types.Timer
	err := r.db.First(&timer, id).Error
	if err != nil {
		return nil, err
	}
	return &timer, nil
}

func (r *timerRepository) Delete(id uint) error {
	return r.db.Delete(&types.Timer{}, id).Error
}

func (r *timerRepository) FindAll() ([]types.Timer, error) {
	var timers []types.Timer
	err := r.db.Find(&timers).Error
	return timers, err
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&types.Timer{})
}
