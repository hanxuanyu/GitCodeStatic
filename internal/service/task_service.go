package service

import (
	"context"

	"github.com/hanxuanyu/gitcodestatic/internal/models"
	"github.com/hanxuanyu/gitcodestatic/internal/storage"
)

// TaskService 任务服务
type TaskService struct {
	store storage.Store
}

// NewTaskService 创建任务服务
func NewTaskService(store storage.Store) *TaskService {
	return &TaskService{
		store: store,
	}
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(ctx context.Context, id int64) (*models.Task, error) {
	return s.store.Tasks().GetByID(ctx, id)
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(ctx context.Context, repoID int64, status string, page, pageSize int) ([]*models.Task, int, error) {
	return s.store.Tasks().List(ctx, repoID, status, page, pageSize)
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(ctx context.Context, id int64) error {
	return s.store.Tasks().Cancel(ctx, id)
}
