package cache

import (
	"testing"

	"github.com/gitcodestatic/gitcodestatic/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestGenerateCacheKey 测试缓存键生成
func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		repoID     int64
		branch     string
		constraint *models.StatsConstraint
		commitHash string
	}{
		{
			name:   "date_range constraint",
			repoID: 1,
			branch: "main",
			constraint: &models.StatsConstraint{
				Type: models.ConstraintTypeDateRange,
				From: "2024-01-01",
				To:   "2024-12-31",
			},
			commitHash: "abc123",
		},
		{
			name:   "commit_limit constraint",
			repoID: 1,
			branch: "main",
			constraint: &models.StatsConstraint{
				Type:  models.ConstraintTypeCommitLimit,
				Limit: 100,
			},
			commitHash: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := GenerateCacheKey(tt.repoID, tt.branch, tt.constraint, tt.commitHash)
			key2 := GenerateCacheKey(tt.repoID, tt.branch, tt.constraint, tt.commitHash)

			// 相同参数应该生成相同的key
			assert.Equal(t, key1, key2)
			assert.NotEmpty(t, key1)
			assert.Len(t, key1, 64) // SHA256 hex = 64 chars
		})
	}

	// 测试不同参数生成不同的key
	t.Run("different parameters generate different keys", func(t *testing.T) {
		constraint := &models.StatsConstraint{
			Type:  models.ConstraintTypeCommitLimit,
			Limit: 100,
		}

		key1 := GenerateCacheKey(1, "main", constraint, "abc123")
		key2 := GenerateCacheKey(1, "main", constraint, "def456") // 不同的commit hash
		key3 := GenerateCacheKey(1, "develop", constraint, "abc123") // 不同的分支

		assert.NotEqual(t, key1, key2)
		assert.NotEqual(t, key1, key3)
		assert.NotEqual(t, key2, key3)
	})
}

// TestSerializeConstraint 测试约束序列化
func TestSerializeConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constraint *models.StatsConstraint
		expected   string
	}{
		{
			name:       "nil constraint",
			constraint: nil,
			expected:   "{}",
		},
		{
			name: "date_range constraint",
			constraint: &models.StatsConstraint{
				Type: models.ConstraintTypeDateRange,
				From: "2024-01-01",
				To:   "2024-12-31",
			},
			expected: `{"type":"date_range","from":"2024-01-01","to":"2024-12-31"}`,
		},
		{
			name: "commit_limit constraint",
			constraint: &models.StatsConstraint{
				Type:  models.ConstraintTypeCommitLimit,
				Limit: 100,
			},
			expected: `{"type":"commit_limit","limit":100}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeConstraint(tt.constraint)
			assert.Equal(t, tt.expected, result)
		})
	}
}
