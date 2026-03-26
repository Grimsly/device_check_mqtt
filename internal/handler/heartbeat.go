package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"device_check_mqtt/internal/model"
	"device_check_mqtt/internal/service"
)

type HeartbeatHandler struct {
	service *service.DeviceStorageService
}

func NewHeartbeatHandler(service *service.DeviceStorageService) *HeartbeatHandler {
	return &HeartbeatHandler{service: service}
}

func (heartbeatHandler *HeartbeatHandler) Post(w http.ResponseWriter, r *http.Request) {
	// Get the wild card device_id from path
	deviceID := r.PathValue("device_id")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req model.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode body", "path", r.URL.Path, "error", err)

		// I recognize that 400 is not defined in the contract, but I don't want to return a 500 for this
		http.Error(w, `{"msg":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.SentAt == nil {
		slog.Error("Request body doesn't contain a required field", "path", r.URL.Path, "missing_field", "sent_at")
		http.Error(w, `{"msg":"Required field sent_at missing"}`, http.StatusBadRequest)
		return
	}

	if err := heartbeatHandler.service.SaveHeartbeat(deviceID, *req.SentAt); err != nil {
		if service.IsNotFound(err) {
			http.Error(w, `{"msg":"Device not found"}`, http.StatusNotFound)
			return
		}
		slog.Error("Failed to save heartbeat", "path", r.URL.Path, "device_id", deviceID, "error", err)
		http.Error(w, `{"msg":"Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
