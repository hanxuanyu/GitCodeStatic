package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gitcodestatic/gitcodestatic/internal/cache"
	"github.com/gitcodestatic/gitcodestatic/internal/git"
	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/models"
	"github.com/gitcodestatic/gitcodestatic/internal/stats"
	"github.com/gitcodestatic/gitcodestatic/internal/storage"
)

// CloneHandler 克隆任务处理器
type CloneHandler struct {
	store      storage.Store
	gitManager git.Manager
}

func NewCloneHandler(store storage.Store, gitManager git.Manager) *CloneHandler {
	return &CloneHandler{
		store:      store,
		gitManager: gitManager,
	}
}

func (h *CloneHandler) Type() string {
	return models.TaskTypeClone
}

func (h *CloneHandler) Timeout() time.Duration {
	return 10 * time.Minute
}

func (h *CloneHandler) Handle(ctx context.Context, task *models.Task) error {
	// 获取仓库信息
	repo, err := h.store.Repos().GetByID(ctx, task.RepoID)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// 更新仓库状态为cloning
	repo.Status = models.RepoStatusCloning
	h.store.Repos().Update(ctx, repo)

	// 获取凭据（如果有）
	var cred *models.Credential
	if repo.CredentialID != nil {
		cred, _ = h.store.Credentials().GetByID(ctx, *repo.CredentialID)
	}

	// 克隆仓库
	if err := h.gitManager.Clone(ctx, repo.URL, repo.LocalPath, cred); err != nil {
		errMsg := err.Error()
		repo.Status = models.RepoStatusFailed
		repo.ErrorMessage = &errMsg
		h.store.Repos().Update(ctx, repo)
		return err
	}

	// 获取当前分支和commit hash
	branch, err := h.gitManager.GetCurrentBranch(ctx, repo.LocalPath)
	if err != nil {
		logger.Logger.Warn().Err(err).Msg("failed to get current branch")
		branch = "main"
	}

	commitHash, err := h.gitManager.GetHeadCommitHash(ctx, repo.LocalPath)
	if err != nil {
		logger.Logger.Warn().Err(err).Msg("failed to get HEAD commit hash")
	}

	// 更新仓库状态为ready
	now := time.Now()
	repo.Status = models.RepoStatusReady
	repo.CurrentBranch = branch
	repo.LastCommitHash = &commitHash
	repo.LastPullAt = &now
	repo.ErrorMessage = nil
	h.store.Repos().Update(ctx, repo)

	return nil
}

// PullHandler 拉取任务处理器
type PullHandler struct {
	store      storage.Store
	gitManager git.Manager
}

func NewPullHandler(store storage.Store, gitManager git.Manager) *PullHandler {
	return &PullHandler{
		store:      store,
		gitManager: gitManager,
	}
}

func (h *PullHandler) Type() string {
	return models.TaskTypePull
}

func (h *PullHandler) Timeout() time.Duration {
	return 5 * time.Minute
}

func (h *PullHandler) Handle(ctx context.Context, task *models.Task) error {
	repo, err := h.store.Repos().GetByID(ctx, task.RepoID)
	if err != nil {
		return err
	}

	var cred *models.Credential
	if repo.CredentialID != nil {
		cred, _ = h.store.Credentials().GetByID(ctx, *repo.CredentialID)
	}

	if err := h.gitManager.Pull(ctx, repo.LocalPath, cred); err != nil {
		return err
	}

	// 更新commit hash
	commitHash, _ := h.gitManager.GetHeadCommitHash(ctx, repo.LocalPath)
	now := time.Now()
	repo.LastCommitHash = &commitHash
	repo.LastPullAt = &now
	h.store.Repos().Update(ctx, repo)

	return nil
}

// SwitchHandler 切换分支处理器
type SwitchHandler struct {
	store      storage.Store
	gitManager git.Manager
}

func NewSwitchHandler(store storage.Store, gitManager git.Manager) *SwitchHandler {
	return &SwitchHandler{
		store:      store,
		gitManager: gitManager,
	}
}

func (h *SwitchHandler) Type() string {
	return models.TaskTypeSwitch
}

func (h *SwitchHandler) Timeout() time.Duration {
	return 1 * time.Minute
}

func (h *SwitchHandler) Handle(ctx context.Context, task *models.Task) error {
	repo, err := h.store.Repos().GetByID(ctx, task.RepoID)
	if err != nil {
		return err
	}

	var params models.TaskParameters
	if err := json.Unmarshal([]byte(task.Parameters), &params); err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	if err := h.gitManager.Checkout(ctx, repo.LocalPath, params.Branch); err != nil {
		return err
	}

	// 更新仓库当前分支
	repo.CurrentBranch = params.Branch
	commitHash, _ := h.gitManager.GetHeadCommitHash(ctx, repo.LocalPath)
	repo.LastCommitHash = &commitHash
	h.store.Repos().Update(ctx, repo)

	return nil
}

// ResetHandler 重置仓库处理器
type ResetHandler struct {
	store      storage.Store
	gitManager git.Manager
	fileCache  *cache.FileCache
}

func NewResetHandler(store storage.Store, gitManager git.Manager, fileCache *cache.FileCache) *ResetHandler {
	return &ResetHandler{
		store:      store,
		gitManager: gitManager,
		fileCache:  fileCache,
	}
}

func (h *ResetHandler) Type() string {
	return models.TaskTypeReset
}

func (h *ResetHandler) Timeout() time.Duration {
	return 10 * time.Minute
}

func (h *ResetHandler) Handle(ctx context.Context, task *models.Task) error {
	repo, err := h.store.Repos().GetByID(ctx, task.RepoID)
	if err != nil {
		return err
	}

	// 1. 删除统计缓存
	h.fileCache.InvalidateByRepoID(ctx, repo.ID)

	// 2. 删除本地目录
	if err := os.RemoveAll(repo.LocalPath); err != nil {
		logger.Logger.Warn().Err(err).Str("path", repo.LocalPath).Msg("failed to remove local path")
	}

	// 3. 更新仓库状态为pending
	repo.Status = models.RepoStatusPending
	repo.CurrentBranch = ""
	repo.LastCommitHash = nil
	repo.LastPullAt = nil
	repo.ErrorMessage = nil
	h.store.Repos().Update(ctx, repo)

	// 4. 重新克隆
	var cred *models.Credential
	if repo.CredentialID != nil {
		cred, _ = h.store.Credentials().GetByID(ctx, *repo.CredentialID)
	}

	repo.Status = models.RepoStatusCloning
	h.store.Repos().Update(ctx, repo)

	if err := h.gitManager.Clone(ctx, repo.URL, repo.LocalPath, cred); err != nil {
		errMsg := err.Error()
		repo.Status = models.RepoStatusFailed
		repo.ErrorMessage = &errMsg
		h.store.Repos().Update(ctx, repo)
		return err
	}

	// 更新为ready
	branch, _ := h.gitManager.GetCurrentBranch(ctx, repo.LocalPath)
	commitHash, _ := h.gitManager.GetHeadCommitHash(ctx, repo.LocalPath)
	now := time.Now()
	repo.Status = models.RepoStatusReady
	repo.CurrentBranch = branch
	repo.LastCommitHash = &commitHash
	repo.LastPullAt = &now
	repo.ErrorMessage = nil
	h.store.Repos().Update(ctx, repo)

	return nil
}

// StatsHandler 统计任务处理器
type StatsHandler struct {
	store      storage.Store
	calculator *stats.Calculator
	fileCache  *cache.FileCache
	gitManager git.Manager
}

func NewStatsHandler(store storage.Store, calculator *stats.Calculator, fileCache *cache.FileCache, gitManager git.Manager) *StatsHandler {
	return &StatsHandler{
		store:      store,
		calculator: calculator,
		fileCache:  fileCache,
		gitManager: gitManager,
	}
}

func (h *StatsHandler) Type() string {
	return models.TaskTypeStats
}

func (h *StatsHandler) Timeout() time.Duration {
	return 30 * time.Minute
}

func (h *StatsHandler) Handle(ctx context.Context, task *models.Task) error {
	repo, err := h.store.Repos().GetByID(ctx, task.RepoID)
	if err != nil {
		return err
	}

	var params models.TaskParameters
	if err := json.Unmarshal([]byte(task.Parameters), &params); err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// 获取当前HEAD commit hash
	commitHash, err := h.gitManager.GetHeadCommitHash(ctx, repo.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to get HEAD commit hash: %w", err)
	}

	// 检查缓存
	cacheKey := cache.GenerateCacheKey(repo.ID, params.Branch, params.Constraint, commitHash)
	cached, _ := h.fileCache.Get(ctx, cacheKey)
	if cached != nil {
		// 缓存命中，直接返回
		logger.Logger.Info().Str("cache_key", cacheKey).Msg("cache hit during stats calculation")
		
		result := models.TaskResult{
			CacheKey: cacheKey,
			Message:  "cache hit",
		}
		resultJSON, _ := json.Marshal(result)
		resultStr := string(resultJSON)
		task.Result = &resultStr
		h.store.Tasks().Update(ctx, task)
		
		return nil
	}

	// 执行统计
	statistics, err := h.calculator.Calculate(ctx, repo.LocalPath, params.Branch, params.Constraint)
	if err != nil {
		return fmt.Errorf("failed to calculate statistics: %w", err)
	}

	// 保存到缓存
	if err := h.fileCache.Set(ctx, repo.ID, params.Branch, params.Constraint, commitHash, statistics); err != nil {
		logger.Logger.Warn().Err(err).Msg("failed to save statistics to cache")
	}

	// 更新任务结果
	result := models.TaskResult{
		CacheKey: cacheKey,
		Message:  "statistics calculated successfully",
	}
	resultJSON, _ := json.Marshal(result)
	resultStr := string(resultJSON)
	task.Result = &resultStr
	h.store.Tasks().Update(ctx, task)

	logger.Logger.Info().
		Int64("repo_id", repo.ID).
		Str("branch", params.Branch).
		Int("total_commits", statistics.Summary.TotalCommits).
		Int("contributors", statistics.Summary.TotalContributors).
		Msg("statistics calculated")

	return nil
}
