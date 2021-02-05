package models

import "time"

type JobExecution struct {
	JobName    string
	Successed  bool
	StartedAt  time.Time
	FinishedAt time.Time
}