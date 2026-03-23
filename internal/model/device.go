package model

import "time"

type HeartbeatStat struct {
	FirstHeartbeat *time.Time
	LastHeartbeat  *time.Time
	HeartbeatCount int
}

type UploadTimeStat struct {
	SentAt     *time.Time
	UploadTime int
}
