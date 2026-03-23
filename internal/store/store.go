package store

import (
	"errors"
	"time"

	"device_check_mqtt/internal/model"
)

// Error output when a device storage operation has failed
var (
	ErrDeviceNotFound      = errors.New("Device not found")
	ErrDeviceAlreadyExists = errors.New("Device already exists")
	ErrDeviceHasNoStats    = errors.New("Device has no stats")
)

// Interface that allows access to device storage operations
// Ensure that all methods properly handles concunrency properly by implementing locks if necessary
type DeviceStorageStore interface {
	// Registers a new device
	AddDevice(id string) error

	// Get the stats given the ID of the device
	GetStats(id string) ([]model.UploadTimeStat, model.HeartbeatStat, error)

	// Record a heartbeat for a device
	SaveHeartbeat(id string, sentAt time.Time) error

	// Record upload time of video for a device
	SaveUploadTime(id string, sentAt time.Time, uploadTime int) error

	// Get all currently stored device IDs
	GetAllDeviceIDs() []string
}
