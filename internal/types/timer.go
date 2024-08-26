package types

type Timer struct {
	ID          uint   `gorm:"primarykey"`
	SessionID   string `gorm:"index"`
	MaxTime     int64
	CurrentTime int64
	IsPaused    bool
}

type TimerRequest struct {
	SessionID string `json:"sessionId"`
	MaxTime   int64  `json:"maxTime"`
}

type TimerResponse struct {
	ID          uint   `json:"id"`
	SessionID   string `json:"sessionId"`
	CurrentTime int64  `json:"currentTime"`
	MaxTime     int64  `json:"maxTime"`
	IsPaused    bool   `json:"isPaused"`
}
