package queue

import "github.com/hibiken/asynq"

const TaskTypeTransaction = "transaction:ingest"

// Enqueuer mirrors the *asynq.Client signature so the real client satisfies
// this interface with zero adapter code.
type Enqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}
