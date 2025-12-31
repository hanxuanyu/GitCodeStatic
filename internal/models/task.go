package models

import "time"

// Task 任务模型
type Task struct {
	ID           int64      `json:"id" db:"id"`
	TaskType     string     `json:"task_type" db:"task_type"`
	RepoID       int64      `json:"repo_id" db:"repo_id"`
	Status       string     `json:"status" db:"status"`
	Priority     int        `json:"priority" db:"priority"`
	Parameters   string     `json:"parameters,omitempty" db:"parameters"`   // JSON string
	Result       *string    `json:"result,omitempty" db:"result"`           // JSON string
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int        `json:"retry_count" db:"retry_count"`
	StartedAt    *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DurationMs   *int64     `json:"duration_ms,omitempty" db:"-"` // 计算字段
}

// Task Type constants
const (
	TaskTypeClone        = "clone"
	TaskTypePull         = "pull"
	TaskTypeSwitch       = "switch"
	TaskTypeReset        = "reset"
	TaskTypeStats        = "stats"
	TaskTypeCountCommits = "count_commits"
)

// Task Status constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

// TaskParameters 任务参数结构
type TaskParameters struct {
	Branch     string              `json:"branch,omitempty"`
	Constraint *StatsConstraint    `json:"constraint,omitempty"`
}

// TaskResult 任务结果结构
type TaskResult struct {
	CacheKey     string `json:"cache_key,omitempty"`
	StatsCacheID int64  `json:"stats_cache_id,omitempty"`
	CommitCount  int    `json:"commit_count,omitempty"`
	Message      string `json:"message,omitempty"`
}
