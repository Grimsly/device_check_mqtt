package model

import "time"

type UploadStatsRequest struct {
	SentAt *time.Time `json:"sent_at,omitzero"`
	// the number of nanoseconds it took to upload a video
	UploadTime *int `json:"upload_time"`
}

type GetDeviceStatsResponse struct {
	// Uptime as a percentage. eg: 98.999
	Uptime float64 `json:"uptime"`
	// returned as a time duration string. Eg: 5m10s
	AvgUploadTime string `json:"avg_upload_time"`
}
