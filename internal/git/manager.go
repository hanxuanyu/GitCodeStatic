package git

import (
	"context"

	"github.com/gitcodestatic/gitcodestatic/internal/models"
)

// Manager Git管理器接口
type Manager interface {
	// Clone 克隆仓库
	Clone(ctx context.Context, url, localPath string, cred *models.Credential) error
	
	// Pull 拉取更新
	Pull(ctx context.Context, localPath string, cred *models.Credential) error
	
	// Checkout 切换分支
	Checkout(ctx context.Context, localPath, branch string) error
	
	// GetCurrentBranch 获取当前分支
	GetCurrentBranch(ctx context.Context, localPath string) (string, error)
	
	// GetHeadCommitHash 获取HEAD commit hash
	GetHeadCommitHash(ctx context.Context, localPath string) (string, error)
	
	// CountCommits 统计提交次数
	CountCommits(ctx context.Context, localPath, branch, fromDate string) (int, error)
	
	// IsAvailable 检查Git是否可用
	IsAvailable() bool
}
