package git

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/models"
)

// CmdGitManager 基于git命令的实现
type CmdGitManager struct {
	gitPath string
}

// NewCmdGitManager 创建命令行Git管理器
func NewCmdGitManager(gitPath string) *CmdGitManager {
	if gitPath == "" {
		gitPath = "git"
	}
	return &CmdGitManager{gitPath: gitPath}
}

// IsAvailable 检查git命令是否可用
func (m *CmdGitManager) IsAvailable() bool {
	cmd := exec.Command(m.gitPath, "--version")
	err := cmd.Run()
	return err == nil
}

// Clone 克隆仓库
func (m *CmdGitManager) Clone(ctx context.Context, url, localPath string, cred *models.Credential) error {
	// 注入凭据到URL（如果有）
	cloneURL := url
	if cred != nil {
		cloneURL = m.injectCredentials(url, cred)
	}

	cmd := exec.CommandContext(ctx, m.gitPath, "clone", cloneURL, localPath)
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0") // 禁止交互式提示

	output, err := cmd.CombinedOutput()
	if err != nil {
		// 脱敏日志
		sanitizedURL := sanitizeURL(url)
		logger.Logger.Error().
			Err(err).
			Str("url", sanitizedURL).
			Str("output", string(output)).
			Msg("failed to clone repository")
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	logger.Logger.Info().
		Str("url", sanitizeURL(url)).
		Str("local_path", localPath).
		Msg("repository cloned successfully")

	return nil
}

// Pull 拉取更新
func (m *CmdGitManager) Pull(ctx context.Context, localPath string, cred *models.Credential) error {
	cmd := exec.CommandContext(ctx, m.gitPath, "-C", localPath, "pull")
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("local_path", localPath).
			Str("output", string(output)).
			Msg("failed to pull repository")
		return fmt.Errorf("failed to pull repository: %w", err)
	}

	logger.Logger.Info().
		Str("local_path", localPath).
		Msg("repository pulled successfully")

	return nil
}

// Checkout 切换分支
func (m *CmdGitManager) Checkout(ctx context.Context, localPath, branch string) error {
	cmd := exec.CommandContext(ctx, m.gitPath, "-C", localPath, "checkout", branch)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("local_path", localPath).
			Str("branch", branch).
			Str("output", string(output)).
			Msg("failed to checkout branch")
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	logger.Logger.Info().
		Str("local_path", localPath).
		Str("branch", branch).
		Msg("branch checked out successfully")

	return nil
}

// GetCurrentBranch 获取当前分支
func (m *CmdGitManager) GetCurrentBranch(ctx context.Context, localPath string) (string, error) {
	cmd := exec.CommandContext(ctx, m.gitPath, "-C", localPath, "rev-parse", "--abbrev-ref", "HEAD")
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// GetHeadCommitHash 获取HEAD commit hash
func (m *CmdGitManager) GetHeadCommitHash(ctx context.Context, localPath string) (string, error) {
	cmd := exec.CommandContext(ctx, m.gitPath, "-C", localPath, "rev-parse", "HEAD")
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit hash: %w", err)
	}

	hash := strings.TrimSpace(string(output))
	return hash, nil
}

// CountCommits 统计提交次数
func (m *CmdGitManager) CountCommits(ctx context.Context, localPath, branch, fromDate string) (int, error) {
	args := []string{"-C", localPath, "rev-list", "--count"}
	
	if fromDate != "" {
		args = append(args, "--since="+fromDate)
	}
	
	args = append(args, branch)
	
	cmd := exec.CommandContext(ctx, m.gitPath, args...)
	
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}

	countStr := strings.TrimSpace(string(output))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count, nil
}

// injectCredentials 注入凭据到URL
func (m *CmdGitManager) injectCredentials(url string, cred *models.Credential) string {
	if cred == nil || cred.Username == "" {
		return url
	}

	// 简单的URL凭据注入（仅支持https）
	if strings.HasPrefix(url, "https://") {
		credentials := cred.Username
		if cred.Password != "" {
			credentials += ":" + cred.Password
		}
		return strings.Replace(url, "https://", "https://"+credentials+"@", 1)
	}

	return url
}

// sanitizeURL 脱敏URL（移除用户名密码）
func sanitizeURL(url string) string {
	re := regexp.MustCompile(`(https?://)[^@]+@`)
	return re.ReplaceAllString(url, "${1}***@")
}
