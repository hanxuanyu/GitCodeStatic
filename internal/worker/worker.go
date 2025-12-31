package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hanxuanyu/gitcodestatic/internal/logger"
	"github.com/hanxuanyu/gitcodestatic/internal/models"
	"github.com/hanxuanyu/gitcodestatic/internal/storage"
)

// TaskHandler 任务处理器接口
type TaskHandler interface {
	Handle(ctx context.Context, task *models.Task) error
	Type() string
	Timeout() time.Duration
}

// Worker 工作器
type Worker struct {
	id       int
	queue    *Queue
	handlers map[string]TaskHandler
	store    storage.Store
	stopCh   chan struct{}
	wg       *sync.WaitGroup
}

// NewWorker 创建工作器
func NewWorker(id int, queue *Queue, store storage.Store, handlers map[string]TaskHandler) *Worker {
	return &Worker{
		id:       id,
		queue:    queue,
		handlers: handlers,
		store:    store,
		stopCh:   make(chan struct{}),
		wg:       &sync.WaitGroup{},
	}
}

// Start 启动工作器
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
}

// Stop 停止工作器
func (w *Worker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

// run 运行工作器
func (w *Worker) run(ctx context.Context) {
	defer w.wg.Done()

	logger.Logger.Info().Int("worker_id", w.id).Msg("worker started")

	for {
		select {
		case <-w.stopCh:
			logger.Logger.Info().Int("worker_id", w.id).Msg("worker stopped")
			return
		case <-ctx.Done():
			logger.Logger.Info().Int("worker_id", w.id).Msg("worker context cancelled")
			return
		default:
			// 从队列取任务
			task, err := w.queue.Dequeue(ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				logger.Logger.Error().Err(err).Int("worker_id", w.id).Msg("failed to dequeue task")
				time.Sleep(time.Second)
				continue
			}

			if task == nil {
				continue
			}

			// 处理任务
			w.handleTask(ctx, task)
		}
	}
}

// handleTask 处理任务
func (w *Worker) handleTask(ctx context.Context, task *models.Task) {
	startTime := time.Now()

	logger.Logger.Info().
		Int("worker_id", w.id).
		Int64("task_id", task.ID).
		Str("task_type", task.TaskType).
		Int64("repo_id", task.RepoID).
		Msg("task started")

	// 更新任务状态为运行中
	if err := w.store.Tasks().UpdateStatus(ctx, task.ID, models.TaskStatusRunning, nil); err != nil {
		logger.Logger.Error().Err(err).Int64("task_id", task.ID).Msg("failed to update task status to running")
		return
	}

	// 查找处理器
	handler, ok := w.handlers[task.TaskType]
	if !ok {
		errMsg := fmt.Sprintf("no handler found for task type: %s", task.TaskType)
		logger.Logger.Error().Int64("task_id", task.ID).Str("task_type", task.TaskType).Msg(errMsg)
		w.store.Tasks().UpdateStatus(ctx, task.ID, models.TaskStatusFailed, &errMsg)
		return
	}

	// 创建带超时的上下文
	timeout := handler.Timeout()
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行任务
	err := handler.Handle(taskCtx, task)

	duration := time.Since(startTime)

	if err != nil {
		errMsg := err.Error()
		logger.Logger.Error().
			Err(err).
			Int("worker_id", w.id).
			Int64("task_id", task.ID).
			Str("task_type", task.TaskType).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("task failed")

		w.store.Tasks().UpdateStatus(ctx, task.ID, models.TaskStatusFailed, &errMsg)
		return
	}

	// 任务成功
	logger.Logger.Info().
		Int("worker_id", w.id).
		Int64("task_id", task.ID).
		Str("task_type", task.TaskType).
		Int64("duration_ms", duration.Milliseconds()).
		Msg("task completed")

	w.store.Tasks().UpdateStatus(ctx, task.ID, models.TaskStatusCompleted, nil)
}
