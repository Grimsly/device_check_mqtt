package router

import (
	"net/http"

	"device_check_mqtt/internal/handler"
	"device_check_mqtt/internal/middleware"
	"device_check_mqtt/internal/service"
)

// Build the router with all the routes
func NewRouter(service *service.DeviceStorageService) http.Handler {
	mux := http.NewServeMux()

	heartbeat := handler.NewHeartbeatHandler(service)
	stats := handler.NewStatsHandler(service)

	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", heartbeat.Post)
	mux.HandleFunc("POST /api/v1/devices/{device_id}/stats", stats.Post)
	mux.HandleFunc("GET /api/v1/devices/{device_id}/stats", stats.Get)

	// Health check
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	var h http.Handler = mux

	h = middleware.RequestLogging(h)

	return h
}
