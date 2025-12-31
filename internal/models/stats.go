package models

import "time"

// StatsCache 统计缓存模型
type StatsCache struct {
	ID              int64     `json:"id" db:"id"`
	RepoID          int64     `json:"repo_id" db:"repo_id"`
	Branch          string    `json:"branch" db:"branch"`
	ConstraintType  string    `json:"constraint_type" db:"constraint_type"`     // date_range/commit_limit
	ConstraintValue string    `json:"constraint_value" db:"constraint_value"`   // JSON string
	CommitHash      string    `json:"commit_hash" db:"commit_hash"`
	ResultPath      string    `json:"result_path" db:"result_path"`
	ResultSize      int64     `json:"result_size" db:"result_size"`
	CacheKey        string    `json:"cache_key" db:"cache_key"`
	HitCount        int       `json:"hit_count" db:"hit_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	LastHitAt       *time.Time `json:"last_hit_at,omitempty" db:"last_hit_at"`
}

// StatsConstraint 统计约束
type StatsConstraint struct {
	Type  string `json:"type"`            // date_range 或 commit_limit
	From  string `json:"from,omitempty"`  // type=date_range时使用
	To    string `json:"to,omitempty"`    // type=date_range时使用
	Limit int    `json:"limit,omitempty"` // type=commit_limit时使用
}

// Constraint Type constants
const (
	ConstraintTypeDateRange   = "date_range"
	ConstraintTypeCommitLimit = "commit_limit"
)

// StatsResult 统计结果
type StatsResult struct {
	CacheHit     bool                 `json:"cache_hit"`
	CachedAt     *time.Time           `json:"cached_at,omitempty"`
	CommitHash   string               `json:"commit_hash"`
	Statistics   *Statistics          `json:"statistics"`
}

// Statistics 统计数据
type Statistics struct {
	Summary         StatsSummary          `json:"summary"`
	ByContributor   []ContributorStats    `json:"by_contributor"`
}

// StatsSummary 统计摘要
type StatsSummary struct {
	TotalCommits      int               `json:"total_commits"`
	TotalContributors int               `json:"total_contributors"`
	DateRange         *DateRange        `json:"date_range,omitempty"`
	CommitLimit       *int              `json:"commit_limit,omitempty"`
}

// DateRange 日期范围
type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ContributorStats 贡献者统计
type ContributorStats struct {
	Author        string `json:"author"`
	Email         string `json:"email"`
	Commits       int    `json:"commits"`
	Additions     int    `json:"additions"`      // 新增行数
	Deletions     int    `json:"deletions"`      // 删除行数
	Modifications int    `json:"modifications"`  // 修改行数 = min(additions, deletions)
	NetAdditions  int    `json:"net_additions"`  // 净增加 = additions - deletions
}

// Credential 凭据模型
type Credential struct {
	ID            string    `json:"id" db:"id"`
	Username      string    `json:"username,omitempty" db:"-"` // 不直接存储，存在EncryptedData中
	Password      string    `json:"password,omitempty" db:"-"` // 不直接存储
	AuthType      string    `json:"auth_type" db:"auth_type"`
	EncryptedData []byte    `json:"-" db:"encrypted_data"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Auth Type constants
const (
	AuthTypeBasic = "basic"
	AuthTypeToken = "token"
	AuthTypeSSH   = "ssh"
)
