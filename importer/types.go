package importer

import "time"

type Status string

const (
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type Job struct {
	SeedID    int        `json:"seed_id"`
	Status    Status     `json:"status"`
	Error     string     `json:"error,omitempty"`
	Entries   int        `json:"entries"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}
