package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hanxuanyu/gitcodestatic/internal/git"
	"github.com/hanxuanyu/gitcodestatic/internal/logger"
	"github.com/hanxuanyu/gitcodestatic/internal/models"
	"github.com/hanxuanyu/gitcodestatic/internal/storage"
	"github.com/hanxuanyu/gitcodestatic/internal/worker"
)

// RepoService 仓库服务
type RepoService struct {
	store      storage.Store
	queue      *worker.Queue
	cacheDir   string
	gitManager git.Manager
}

// NewRepoService 创建仓库服务
func NewRepoService(store storage.Store, queue *worker.Queue, cacheDir string, gitManager git.Manager) *RepoService {
	return &RepoService{
		store:      store,
		queue:      queue,
		cacheDir:   cacheDir,
		gitManager: gitManager,
	}
}

// AddReposRequest 批量添加仓库请求
// RepoInput 仓库输入
type RepoInput struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
}

type AddReposRequest struct {
	Repos    []RepoInput `json:"repos"`
	Username string      `json:"username,omitempty"` // 可选的认证信息
	Password string      `json:"password,omitempty"` // 可选的认证信息
}

// AddReposResponse 批量添加仓库响应
type AddReposResponse struct {
	Total     int              `json:"total"`
	Succeeded []AddRepoResult  `json:"succeeded"`
	Failed    []AddRepoFailure `json:"failed"`
}

// AddRepoResult 添加仓库成功结果
type AddRepoResult struct {
	RepoID int64  `json:"repo_id"`
	URL    string `json:"url"`
	TaskID int64  `json:"task_id"`
}

// AddRepoFailure 添加仓库失败结果
type AddRepoFailure struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

// AddRepos 批量添加仓库
func (s *RepoService) AddRepos(ctx context.Context, req *AddReposRequest) (*AddReposResponse, error) {
	resp := &AddReposResponse{
		Total:     len(req.Repos),
		Succeeded: make([]AddRepoResult, 0),
		Failed:    make([]AddRepoFailure, 0),
	}

	// 如果提供了认证信息，创建凭据
	var credentialID *string
	if req.Username != "" && req.Password != "" {
		cred := &models.Credential{
			ID:       generateCredentialID(),
			Username: req.Username,
			Password: req.Password,
			AuthType: models.AuthTypeBasic,
		}

		if err := s.store.Credentials().Create(ctx, cred); err != nil {
			logger.Logger.Warn().Err(err).Msg("failed to save credential, will continue without credentials")
		} else {
			credentialID = &cred.ID
			logger.Logger.Info().Str("credential_id", cred.ID).Msg("credential created")
		}
	}

	for _, repoInput := range req.Repos {
		url := repoInput.URL
		branch := repoInput.Branch
		if branch == "" {
			branch = "main" // 默认分支
		}

		// 校验URL
		if !isValidGitURL(url) {
			resp.Failed = append(resp.Failed, AddRepoFailure{
				URL:   url,
				Error: "invalid git URL",
			})
			continue
		}

		// 检查是否已存在
		existing, err := s.store.Repos().GetByURL(ctx, url)
		if err != nil {
			resp.Failed = append(resp.Failed, AddRepoFailure{
				URL:   url,
				Error: fmt.Sprintf("failed to check existing repo: %v", err),
			})
			continue
		}

		if existing != nil {
			resp.Failed = append(resp.Failed, AddRepoFailure{
				URL:   url,
				Error: "repository already exists",
			})
			continue
		}

		// 创建仓库记录
		repoName := extractRepoName(url)
		localPath := filepath.Join(s.cacheDir, repoName)

		repo := &models.Repository{
			URL:           url,
			Name:          repoName,
			CurrentBranch: branch,
			LocalPath:     localPath,
			Status:        models.RepoStatusPending,
			CredentialID:  credentialID,
		}

		if err := s.store.Repos().Create(ctx, repo); err != nil {
			resp.Failed = append(resp.Failed, AddRepoFailure{
				URL:   url,
				Error: fmt.Sprintf("failed to create repository: %v", err),
			})
			continue
		}

		// 提交clone任务
		task := &models.Task{
			TaskType: models.TaskTypeClone,
			RepoID:   repo.ID,
			Priority: 0,
		}

		if err := s.queue.Enqueue(ctx, task); err != nil {
			resp.Failed = append(resp.Failed, AddRepoFailure{
				URL:   url,
				Error: fmt.Sprintf("failed to enqueue clone task: %v", err),
			})
			continue
		}

		resp.Succeeded = append(resp.Succeeded, AddRepoResult{
			RepoID: repo.ID,
			URL:    url,
			TaskID: task.ID,
		})

		logger.Logger.Info().
			Int64("repo_id", repo.ID).
			Str("url", url).
			Int64("task_id", task.ID).
			Bool("has_credentials", credentialID != nil).
			Msg("repository added")
	}

	return resp, nil
}

// GetRepo 获取仓库详情
func (s *RepoService) GetRepo(ctx context.Context, id int64) (*models.Repository, error) {
	return s.store.Repos().GetByID(ctx, id)
}

// ListRepos 获取仓库列表
func (s *RepoService) ListRepos(ctx context.Context, status string, page, pageSize int) ([]*models.Repository, int, error) {
	repos, total, err := s.store.Repos().List(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 设置has_credentials标志
	for _, repo := range repos {
		repo.HasCredentials = repo.CredentialID != nil && *repo.CredentialID != ""
	}

	return repos, total, nil
}

// SwitchBranch 切换分支
func (s *RepoService) SwitchBranch(ctx context.Context, repoID int64, branch string) (*models.Task, error) {
	// 检查仓库是否存在
	repo, err := s.store.Repos().GetByID(ctx, repoID)
	if err != nil {
		return nil, err
	}

	if repo.Status != models.RepoStatusReady {
		return nil, errors.New("repository is not ready")
	}

	// 创建切换分支任务
	params := models.TaskParameters{
		Branch: branch,
	}
	paramsJSON, _ := json.Marshal(params)

	task := &models.Task{
		TaskType:   models.TaskTypeSwitch,
		RepoID:     repoID,
		Parameters: string(paramsJSON),
		Priority:   0,
	}

	if err := s.queue.Enqueue(ctx, task); err != nil {
		return nil, err
	}

	logger.Logger.Info().
		Int64("repo_id", repoID).
		Str("branch", branch).
		Int64("task_id", task.ID).
		Msg("switch branch task submitted")

	return task, nil
}

// UpdateRepo 更新仓库（pull）
func (s *RepoService) UpdateRepo(ctx context.Context, repoID int64) (*models.Task, error) {
	// 检查仓库是否存在
	repo, err := s.store.Repos().GetByID(ctx, repoID)
	if err != nil {
		return nil, err
	}

	if repo.Status != models.RepoStatusReady {
		return nil, errors.New("repository is not ready")
	}

	// 创建pull任务
	task := &models.Task{
		TaskType: models.TaskTypePull,
		RepoID:   repoID,
		Priority: 0,
	}

	if err := s.queue.Enqueue(ctx, task); err != nil {
		return nil, err
	}

	logger.Logger.Info().
		Int64("repo_id", repoID).
		Int64("task_id", task.ID).
		Msg("update task submitted")

	return task, nil
}

// ResetRepo 重置仓库
func (s *RepoService) ResetRepo(ctx context.Context, repoID int64) (*models.Task, error) {
	// 检查仓库是否存在
	_, err := s.store.Repos().GetByID(ctx, repoID)
	if err != nil {
		return nil, err
	}

	// 创建reset任务
	task := &models.Task{
		TaskType: models.TaskTypeReset,
		RepoID:   repoID,
		Priority: 1, // 高优先级
	}

	if err := s.queue.Enqueue(ctx, task); err != nil {
		return nil, err
	}

	logger.Logger.Info().
		Int64("repo_id", repoID).
		Int64("task_id", task.ID).
		Msg("reset task submitted")

	return task, nil
}

// DeleteRepo 删除仓库
func (s *RepoService) DeleteRepo(ctx context.Context, id int64) error {
	return s.store.Repos().Delete(ctx, id)
}

// GetBranches 获取仓库分支列表
func (s *RepoService) GetBranches(ctx context.Context, repoID int64) ([]string, error) {
	// 获取仓库信息
	repo, err := s.store.Repos().GetByID(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	if repo == nil {
		return nil, fmt.Errorf("repository not found")
	}

	if repo.Status != models.RepoStatusReady {
		return nil, fmt.Errorf("repository is not ready, status: %s", repo.Status)
	}

	// 使用git命令获取分支列表
	branches, err := s.gitManager.ListBranches(ctx, repo.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	return branches, nil
}

// isValidGitURL 校验Git URL
func isValidGitURL(url string) bool {
	// 简单校验：https:// 或 git@ 开头
	return strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "git@")
}

// extractRepoName 从URL提取仓库名称
func extractRepoName(url string) string {
	// 移除.git后缀
	url = strings.TrimSuffix(url, ".git")

	// 提取最后一个路径部分
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		// 移除特殊字符
		name = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(name, "_")
		return name
	}

	return "repo"
}

// generateCredentialID 生成凭据ID
func generateCredentialID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
