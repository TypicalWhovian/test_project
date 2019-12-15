package db

import "time"

const (
	STATUSRUNNING  = "RUNNING" // when task runs 10 1-second synchronous functions
	STATUSSTOPPED  = "STOPPED" // when task was stopped via request
	STATUSFINISHED = "FINISHED" // when task has ran all 10 1-second synchronous functions
)

type Task struct {
	Id             string    `pg:"id"`
	RequestId      string    `pg:"request_id"`
	StepsCompleted int64     `pg:"steps_completed,use_zero"`
	Status         string    `pg:"status"`
	CreatedAt      time.Time `pg:"created_at"`
	UpdatedAt      time.Time `pg:"updated_at"`
}
