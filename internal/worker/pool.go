package worker

import (
	"context"
	"sync"

	"github.com/hanxuanyu/gitcodestatic/internal/logger"
	"github.com/hanxuanyu/gitcodestatic/internal/storage"
)

// Pool Worker池
type Pool struct {
	queue    *Queue
	workers  []*Worker
	handlers map[string]TaskHandler
	store    storage.Store
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewPool 创建Worker池
func NewPool(workerCount int, queue *Queue, store storage.Store, handlers map[string]TaskHandler) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		queue:    queue,
		workers:  make([]*Worker, 0, workerCount),
		handlers: handlers,
		store:    store,
		ctx:      ctx,
		cancel:   cancel,
	}

	// 创建workers
	for i := 0; i < workerCount; i++ {
		worker := NewWorker(i+1, queue, store, handlers)
		pool.workers = append(pool.workers, worker)
	}

	return pool
}

// Start 启动Worker池
func (p *Pool) Start() {
	logger.Logger.Info().Int("worker_count", len(p.workers)).Msg("starting worker pool")

	for _, worker := range p.workers {
		worker.Start(p.ctx)
	}
}

// Stop 停止Worker池
func (p *Pool) Stop() {
	logger.Logger.Info().Msg("stopping worker pool")

	p.cancel()

	for _, worker := range p.workers {
		worker.Stop()
	}

	p.queue.Close()

	logger.Logger.Info().Msg("worker pool stopped")
}

// GetQueue 获取队列
func (p *Pool) GetQueue() *Queue {
	return p.queue
}

// QueueSize 获取队列长度
func (p *Pool) QueueSize() int {
	return p.queue.Size()
}
