package model

import "time"

type HeartbeatRequest struct {
	SentAt *time.Time `json:"sent_at,omitzero"`
}
