package types

import "encoding/json"

type MessageType string

const (
	TypeTimerUpdate MessageType = "TIMER_UPDATE"
	TypeTimerCreate MessageType = "TIMER_CREATE"
	TypeTimerPause  MessageType = "TIMER_PAUSE"
	TypeTimerResume MessageType = "TIMER_RESUME"
	TypeTimerStop   MessageType = "TIMER_STOP"
	TypeTimerModify MessageType = "TIMER_MODIFY"
)

type WebSocketMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
