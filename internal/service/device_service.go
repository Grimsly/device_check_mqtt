package service

import (
	"errors"
	"log/slog"
	"time"

	"device_check_mqtt/internal/model"
	"device_check_mqtt/internal/store"
)

type DeviceStorageService struct {
	store store.DeviceStorageStore
}

var (
	ErrHeartbeatFirstLastError = errors.New("Device's first heartbeat is later than the last heartbeat")
)

// Initialize a device service
func NewDeviceService(s store.DeviceStorageStore) *DeviceStorageService {
	return &DeviceStorageService{store: s}
}

// Save a heartbeat to the device store
func (service *DeviceStorageService) SaveHeartbeat(deviceID string, sentAt time.Time) error {
	if sentAt.IsZero() {
		return errors.New("sent_at cannot not be zero")
	}
	return service.store.SaveHeartbeat(deviceID, sentAt)
}

// Save uploaded time to the device store
func (service *DeviceStorageService) SaveUploadTime(deviceID string, sentAt time.Time, uploadTime int) error {
	// Used to check and reject sent_at if 0, but I got the device simulator sends 0s
	// if sentAt.IsZero() {
	// 	return errors.New("sent_at cannot not be zero")
	// }

	if uploadTime < 0 {
		return errors.New("upload_time cannot not be negative")
	}

	return service.store.SaveUploadTime(deviceID, sentAt, uploadTime)
}

// Get a device's stats like uptime and average upload time
func (service *DeviceStorageService) GetDeviceStats(deviceID string) (model.GetDeviceStatsResponse, error) {
	uploadTimeStats, heartbeatStat, err := service.store.GetStats(deviceID)
	if err != nil {
		return model.GetDeviceStatsResponse{}, err
	}

	if len(uploadTimeStats) == 0 && heartbeatStat.HeartbeatCount == 0 {
		return model.GetDeviceStatsResponse{Uptime: 0,
			AvgUploadTime: "0s"}, store.ErrDeviceHasNoStats
	}

	// Not sure how to handle errors when the heartbeat comparison for LastHeartbeat and FirstHeartbeat fails.
	// I don't want to just return an error because avgUpload could be a viable stat, but uptime is not.
	// Either I do simple logging like I have here,
	// or I create a Logging service that DeviceStorageService takes in so that I can send the error logs to the cloud or written in a log file
	uptime, err := service.calculateUptime(heartbeatStat)

	if err != nil {
		slog.Error("Heartbeat handling failure", "error", err)
	}

	avgUpload := service.calculateAvgUploadTime(uploadTimeStats)

	return model.GetDeviceStatsResponse{
		Uptime:        uptime,
		AvgUploadTime: avgUpload,
	}, nil
}

// Calculate total uptime of a device given heartbeat stats
//
// Returns 0 if no heartbeats
func (service *DeviceStorageService) calculateUptime(heartbeatStat model.HeartbeatStat) (float64, error) {
	if heartbeatStat.FirstHeartbeat == nil || heartbeatStat.LastHeartbeat == nil {
		return 0, nil
	}

	minutes := heartbeatStat.LastHeartbeat.Sub(*heartbeatStat.FirstHeartbeat).Minutes()
	// Check to make sure that the difference between the LastHeartbeat and the FirstHeartbeat isn't 0 or below it
	if minutes == 0 {
		// If the difference is 0, that could mean that he LastHeartbeat and the Firstbeat are the same, so return the heartbeat count
		minutes = float64(heartbeatStat.HeartbeatCount)
	} else if minutes < 0 {
		// If the difference is less that 0, that means that the LastHeartbeat is earlier than the FirstHeartbeat,
		// which means it could be an issue with the StoreService's method's query
		return -1, ErrHeartbeatFirstLastError
	}

	// uptime = (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
	return (float64(heartbeatStat.HeartbeatCount) / minutes) * 100, nil
}

// Calculate average upload time given a list of upload time stats
//
// Returns "0s" if no uploaded times
func (service *DeviceStorageService) calculateAvgUploadTime(uploadTimeStats []model.UploadTimeStat) string {
	if len(uploadTimeStats) == 0 {
		return "0s"
	}

	sum := 0
	for _, timeStat := range uploadTimeStats {
		sum += timeStat.UploadTime
	}

	// timeDuration = avg(arrayOfUploadTimeDurations)
	avg := float64(sum) / float64(len(uploadTimeStats))
	return time.Duration(avg).String()
}

// Checks if an error is a Device not found error
func IsNotFound(err error) bool {
	return errors.Is(err, store.ErrDeviceNotFound)
}

// Checks if an error is a Device has no stats error
func HasNoStats(err error) bool {
	return errors.Is(err, store.ErrDeviceHasNoStats)
}
