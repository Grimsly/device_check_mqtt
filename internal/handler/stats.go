package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"device_check_mqtt/internal/model"
	"device_check_mqtt/internal/service"
)

type StatsHandler struct {
	service *service.DeviceStorageService
}

func NewStatsHandler(service *service.DeviceStorageService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (statsHandler *StatsHandler) Post(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req model.UploadStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode body", "path", r.URL.Path, "error", err)
		http.Error(w, `{"msg":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.SentAt == nil {
		slog.Error("Request body doesn't contain a required field", "path", r.URL.Path, "missing_field", "sent_at")
		http.Error(w, `{"msg":"Required field sent_at missing"}`, http.StatusBadRequest)
		return
	}

	if req.UploadTime == nil {
		slog.Error("Request body doesn't contain a required field", "path", r.URL.Path, "missing_field", "upload_time")
		http.Error(w, `{"msg":"Required field upload_time missing"}`, http.StatusBadRequest)
		return
	}

	if err := statsHandler.service.SaveUploadTime(deviceID, *req.SentAt, *req.UploadTime); err != nil {
		if service.IsNotFound(err) {
			http.Error(w, `{"msg":"Device not found"}`, http.StatusNotFound)
			return
		}
		slog.Error("Failed to save upload time", "path", r.URL.Path, "device_id", deviceID, "error", err)
		http.Error(w, `{"msg":"Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (statsHandler *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")

	stats, err := statsHandler.service.GetDeviceStats(deviceID)
	if err != nil {
		if service.IsNotFound(err) {
			http.Error(w, `{"msg":"Device not found"}`, http.StatusNotFound)
			return
		}

		if service.HasNoStats(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		slog.Error("Failed to get stats", "path", r.URL.Path, "device_id", deviceID, "error", err)
		http.Error(w, `{"msg":"Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("Failed to encode stats response", "path", r.URL.Path, "device_id", deviceID, "error", err)
	}
}
