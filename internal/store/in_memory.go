package store

import (
	"maps"
	"slices"
	"sync"
	"time"

	"device_check_mqtt/internal/model"
)

// Device in the fleet. Only used in in-memory store.
type Device struct {
	ID             string
	FirstHeartbeat *time.Time
	LastHeartbeat  *time.Time
	HeartbeatCount int
	UploadTimes    []model.UploadTimeStat
}

// In-memory store contains the mutex to lock read and write if needed and a map of stored devices
type InMemoryStore struct {
	mu      sync.RWMutex
	devices map[string]*Device
}

// Create an in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		// Set a size of 100 for the map just in case
		devices: make(map[string]*Device, 100),
	}
}

// Registers a device into store
func (s *InMemoryStore) AddDevice(id string) error {
	// As memory storage is not exactly safe for concurrent use, I implemented mutex locks
	// If a database store is used instead, the database would most likely be able to handle concurrency
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[id]; exists {
		return ErrDeviceAlreadyExists
	}
	s.devices[id] = &Device{ID: id}
	return nil
}

// Get stats from a device
func (s *InMemoryStore) GetStats(id string) ([]model.UploadTimeStat, model.HeartbeatStat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, ok := s.devices[id]
	if !ok {
		return []model.UploadTimeStat{}, model.HeartbeatStat{}, ErrDeviceNotFound
	}

	uploadTimeStat := s.getUploadTimeStat(device)
	heartbeatStat := s.getHeartbeatStat(device)

	return uploadTimeStat, heartbeatStat, nil
}

// Save/record a heartbeat for a device
func (s *InMemoryStore) SaveHeartbeat(id string, sentAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, ok := s.devices[id]
	if !ok {
		return ErrDeviceNotFound
	}

	device.HeartbeatCount++

	// Heartbeat could come in out of order, so ensure that the proper first and last heartbeat is saved
	if device.FirstHeartbeat == nil || sentAt.Before(*device.FirstHeartbeat) {
		device.FirstHeartbeat = new(sentAt)
	}
	if device.LastHeartbeat == nil || sentAt.After(*device.LastHeartbeat) {
		device.LastHeartbeat = new(sentAt)
	}

	return nil
}

// Save device's upload time
func (s *InMemoryStore) SaveUploadTime(id string, sentAt time.Time, uploadTime int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, ok := s.devices[id]
	if !ok {
		return ErrDeviceNotFound
	}

	device.UploadTimes = append(device.UploadTimes, model.UploadTimeStat{SentAt: new(sentAt), UploadTime: uploadTime})

	return nil
}

// Get all device IDs
func (s *InMemoryStore) GetAllDeviceIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return slices.Collect(maps.Keys(s.devices))
}

func (s *InMemoryStore) getUploadTimeStat(device *Device) []model.UploadTimeStat {
	return slices.Clone(device.UploadTimes)
}

func (s *InMemoryStore) getHeartbeatStat(device *Device) model.HeartbeatStat {
	stat := model.HeartbeatStat{
		HeartbeatCount: device.HeartbeatCount,
	}

	if device.FirstHeartbeat != nil {
		stat.FirstHeartbeat = new(*device.FirstHeartbeat)
	}
	if device.LastHeartbeat != nil {
		stat.LastHeartbeat = new(*device.LastHeartbeat)
	}

	return stat
}
