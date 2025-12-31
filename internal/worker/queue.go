package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/hanxuanyu/gitcodestatic/internal/logger"
	"github.com/hanxuanyu/gitcodestatic/internal/models"
	"github.com/hanxuanyu/gitcodestatic/internal/storage"
)

// Queue 任务队列
type Queue struct {
	taskChan chan *models.Task
	store    storage.Store
	mu       sync.RWMutex
}

// NewQueue 创建任务队列
func NewQueue(bufferSize int, store storage.Store) *Queue {
	return &Queue{
		taskChan: make(chan *models.Task, bufferSize),
		store:    store,
	}
}

// Enqueue 加入任务到队列
func (q *Queue) Enqueue(ctx context.Context, task *models.Task) error {
	// 检查是否存在相同的待处理任务（去重）
	existing, err := q.store.Tasks().FindExisting(ctx, task.RepoID, task.TaskType, task.Parameters)
	if err != nil {
		return fmt.Errorf("failed to check existing task: %w", err)
	}

	if existing != nil {
		// 已存在相同任务，返回已有任务
		logger.Logger.Info().
			Int64("task_id", existing.ID).
			Int64("repo_id", task.RepoID).
			Str("task_type", task.TaskType).
			Msg("task already exists, returning existing task")

		task.ID = existing.ID
		task.Status = existing.Status
		task.CreatedAt = existing.CreatedAt
		return nil
	}

	// 创建新任务
	task.Status = models.TaskStatusPending
	if err := q.store.Tasks().Create(ctx, task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// 加入队列
	select {
	case q.taskChan <- task:
		logger.Logger.Info().
			Int64("task_id", task.ID).
			Int64("repo_id", task.RepoID).
			Str("task_type", task.TaskType).
			Msg("task enqueued")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Dequeue 从队列取出任务
func (q *Queue) Dequeue(ctx context.Context) (*models.Task, error) {
	select {
	case task := <-q.taskChan:
		return task, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Size 返回队列长度
func (q *Queue) Size() int {
	return len(q.taskChan)
}

// Close 关闭队列
func (q *Queue) Close() {
	close(q.taskChan)
}
