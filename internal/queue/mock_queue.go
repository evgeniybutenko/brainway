package queue

import "github.com/hibiken/asynq"

// MockEnqueuer is a test double for Enqueuer.
type MockEnqueuer struct {
	// EnqueueFunc lets individual tests inject custom behaviour (e.g. errors).
	EnqueueFunc func(*asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error)
	// Calls records every task passed to Enqueue for assertion in tests.
	Calls []*asynq.Task
}

func (m *MockEnqueuer) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	m.Calls = append(m.Calls, task)
	if m.EnqueueFunc != nil {
		return m.EnqueueFunc(task, opts...)
	}
	return &asynq.TaskInfo{}, nil
}
