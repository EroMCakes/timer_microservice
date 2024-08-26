package server

import (
	"timer-microservice/internal/handlers"
	"timer-microservice/internal/websocket"
)

func (s *Server) SetupRoutes(th *handlers.TimerHandler, wsh *websocket.Handler) {
	s.router.Post("/timer", th.CreateTimer)
	s.router.Put("/timer/{id}/pause", th.PauseTimer)
	s.router.Put("/timer/{id}/resume", th.ResumeTimer)
	s.router.Put("/timer/{id}/stop", th.StopTimer)
	s.router.Put("/timer/{id}/modify", th.ModifyTimer)
	s.router.Get("/ws/customer/{sessionID}", wsh.HandleCustomerWebSocket)
	s.router.Get("/ws/gamemaster/{sessionID}", wsh.HandleGameMasterWebSocket)
}
