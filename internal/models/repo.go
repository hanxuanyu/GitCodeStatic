package models

import "time"

// Repository 仓库模型
type Repository struct {
	ID             int64      `json:"id" db:"id"`
	URL            string     `json:"url" db:"url"`
	Name           string     `json:"name" db:"name"`
	CurrentBranch  string     `json:"current_branch" db:"current_branch"`
	LocalPath      string     `json:"local_path" db:"local_path"`
	Status         string     `json:"status" db:"status"` // pending/cloning/ready/failed
	ErrorMessage   *string    `json:"error_message,omitempty" db:"error_message"`
	LastPullAt     *time.Time `json:"last_pull_at,omitempty" db:"last_pull_at"`
	LastCommitHash *string    `json:"last_commit_hash,omitempty" db:"last_commit_hash"`
	CredentialID   *string    `json:"-" db:"credential_id"` // 不返回给前端
	HasCredentials bool       `json:"has_credentials" db:"-"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Repository Status constants
const (
	RepoStatusPending = "pending"
	RepoStatusCloning = "cloning"
	RepoStatusReady   = "ready"
	RepoStatusFailed  = "failed"
)
