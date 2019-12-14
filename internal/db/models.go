package db

import "time"

const (
	STATUSRUNNING  = "RUNNING"
	STATUSSTOPPED  = "STOPPED"
	STATUSFINISHED = "FINISHED"
)

type Task struct {
	Id             string    `pg:"id"`
	RequestId      string    `pg:"request_id"`
	StepsCompleted int64     `pg:"steps_completed"`
	Status         string    `pg:"status"`
	CreatedAt      time.Time `pg:"created_at"`
	UpdatedAt      time.Time `pg:"updated_at"`
}
