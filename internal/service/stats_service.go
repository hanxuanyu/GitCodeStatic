package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hanxuanyu/gitcodestatic/internal/cache"
	"github.com/hanxuanyu/gitcodestatic/internal/git"
	"github.com/hanxuanyu/gitcodestatic/internal/logger"
	"github.com/hanxuanyu/gitcodestatic/internal/models"
	"github.com/hanxuanyu/gitcodestatic/internal/storage"
	"github.com/hanxuanyu/gitcodestatic/internal/worker"
)

// StatsService 统计服务
type StatsService struct {
	store      storage.Store
	queue      *worker.Queue
	cache      *cache.FileCache
	gitManager git.Manager
}

// NewStatsService 创建统计服务
func NewStatsService(store storage.Store, queue *worker.Queue, fileCache *cache.FileCache, gitManager git.Manager) *StatsService {
	return &StatsService{
		store:      store,
		queue:      queue,
		cache:      fileCache,
		gitManager: gitManager,
	}
}

// CalculateRequest 统计请求
type CalculateRequest struct {
	RepoID     int64                   `json:"repo_id"`
	Branch     string                  `json:"branch"`
	Constraint *models.StatsConstraint `json:"constraint"`
}

// Calculate 触发统计计算
func (s *StatsService) Calculate(ctx context.Context, req *CalculateRequest) (*models.Task, error) {
	// 校验参数
	if err := ValidateStatsConstraint(req.Constraint); err != nil {
		return nil, err
	}

	// 检查仓库
	repo, err := s.store.Repos().GetByID(ctx, req.RepoID)
	if err != nil {
		return nil, err
	}

	if repo.Status != models.RepoStatusReady {
		return nil, errors.New("repository is not ready")
	}

	// 创建统计任务
	params := models.TaskParameters{
		Branch:     req.Branch,
		Constraint: req.Constraint,
	}
	paramsJSON, _ := json.Marshal(params)

	task := &models.Task{
		TaskType:   models.TaskTypeStats,
		RepoID:     req.RepoID,
		Parameters: string(paramsJSON),
		Priority:   0,
	}

	if err := s.queue.Enqueue(ctx, task); err != nil {
		return nil, err
	}

	logger.Logger.Info().
		Int64("repo_id", req.RepoID).
		Str("branch", req.Branch).
		Int64("task_id", task.ID).
		Msg("stats task submitted")

	return task, nil
}

// QueryResultRequest 查询统计结果请求
type QueryResultRequest struct {
	RepoID         int64  `json:"repo_id"`
	Branch         string `json:"branch"`
	ConstraintType string `json:"constraint_type"`
	From           string `json:"from,omitempty"`
	To             string `json:"to,omitempty"`
	Limit          int    `json:"limit,omitempty"`
}

// QueryResult 查询统计结果
func (s *StatsService) QueryResult(ctx context.Context, req *QueryResultRequest) (*models.StatsResult, error) {
	// 检查仓库
	repo, err := s.store.Repos().GetByID(ctx, req.RepoID)
	if err != nil {
		return nil, err
	}

	if repo.Status != models.RepoStatusReady {
		return nil, errors.New("repository is not ready")
	}

	// 构建约束
	constraint := &models.StatsConstraint{
		Type: req.ConstraintType,
	}
	if req.ConstraintType == models.ConstraintTypeDateRange {
		constraint.From = req.From
		constraint.To = req.To
	} else {
		constraint.Limit = req.Limit
	}

	// 获取当前HEAD commit hash
	commitHash, err := s.gitManager.GetHeadCommitHash(ctx, repo.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit hash: %w", err)
	}

	// 生成缓存键
	cacheKey := cache.GenerateCacheKey(req.RepoID, req.Branch, constraint, commitHash)

	// 查询缓存
	result, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		logger.Logger.Warn().Err(err).Str("cache_key", cacheKey).Msg("failed to get cache")
	}

	if result != nil {
		return result, nil
	}

	// 缓存未命中
	return nil, errors.New("statistics not found, please submit calculation task first")
}

// CountCommitsRequest 统计提交次数请求
type CountCommitsRequest struct {
	RepoID int64  `json:"repo_id"`
	Branch string `json:"branch"`
	From   string `json:"from"`
}

// CountCommitsResponse 统计提交次数响应
type CountCommitsResponse struct {
	RepoID      int64  `json:"repo_id"`
	Branch      string `json:"branch"`
	From        string `json:"from"`
	To          string `json:"to"`
	CommitCount int    `json:"commit_count"`
}

// CountCommits 统计提交次数（辅助查询）
func (s *StatsService) CountCommits(ctx context.Context, req *CountCommitsRequest) (*CountCommitsResponse, error) {
	// 检查仓库
	repo, err := s.store.Repos().GetByID(ctx, req.RepoID)
	if err != nil {
		return nil, err
	}

	if repo.Status != models.RepoStatusReady {
		return nil, errors.New("repository is not ready")
	}

	// 统计提交次数
	count, err := s.gitManager.CountCommits(ctx, repo.LocalPath, req.Branch, req.From)
	if err != nil {
		return nil, fmt.Errorf("failed to count commits: %w", err)
	}

	resp := &CountCommitsResponse{
		RepoID:      req.RepoID,
		Branch:      req.Branch,
		From:        req.From,
		To:          "HEAD",
		CommitCount: count,
	}

	logger.Logger.Info().
		Int64("repo_id", req.RepoID).
		Str("branch", req.Branch).
		Str("from", req.From).
		Int("count", count).
		Msg("commits counted")

	return resp, nil
}

// ValidateStatsConstraint 校验统计约束
func ValidateStatsConstraint(constraint *models.StatsConstraint) error {
	if constraint == nil {
		return errors.New("constraint is required")
	}

	if constraint.Type != models.ConstraintTypeDateRange && constraint.Type != models.ConstraintTypeCommitLimit {
		return fmt.Errorf("constraint type must be %s or %s", models.ConstraintTypeDateRange, models.ConstraintTypeCommitLimit)
	}

	if constraint.Type == models.ConstraintTypeDateRange {
		if constraint.From == "" || constraint.To == "" {
			return fmt.Errorf("%s requires both from and to", models.ConstraintTypeDateRange)
		}
		if constraint.Limit != 0 {
			return fmt.Errorf("%s cannot be used with limit", models.ConstraintTypeDateRange)
		}
	} else if constraint.Type == models.ConstraintTypeCommitLimit {
		if constraint.Limit <= 0 {
			return fmt.Errorf("%s requires positive limit value", models.ConstraintTypeCommitLimit)
		}
		if constraint.From != "" || constraint.To != "" {
			return fmt.Errorf("%s cannot be used with date range", models.ConstraintTypeCommitLimit)
		}
	}

	return nil
}
