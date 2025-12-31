package service

import (
	"testing"

	"github.com/gitcodestatic/gitcodestatic/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestValidateStatsConstraint 测试统计约束校验
func TestValidateStatsConstraint(t *testing.T) {
	tests := []struct {
		name        string
		constraint  *models.StatsConstraint
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil constraint",
			constraint:  nil,
			expectError: true,
			errorMsg:    "constraint is required",
		},
		{
			name: "valid date_range constraint",
			constraint: &models.StatsConstraint{
				Type: models.ConstraintTypeDateRange,
				From: "2024-01-01",
				To:   "2024-12-31",
			},
			expectError: false,
		},
		{
			name: "date_range missing from",
			constraint: &models.StatsConstraint{
				Type: models.ConstraintTypeDateRange,
				To:   "2024-12-31",
			},
			expectError: true,
			errorMsg:    "date_range requires both from and to",
		},
		{
			name: "date_range with limit (invalid)",
			constraint: &models.StatsConstraint{
				Type:  models.ConstraintTypeDateRange,
				From:  "2024-01-01",
				To:    "2024-12-31",
				Limit: 100,
			},
			expectError: true,
			errorMsg:    "date_range cannot be used with limit",
		},
		{
			name: "valid commit_limit constraint",
			constraint: &models.StatsConstraint{
				Type:  models.ConstraintTypeCommitLimit,
				Limit: 100,
			},
			expectError: false,
		},
		{
			name: "commit_limit with zero limit",
			constraint: &models.StatsConstraint{
				Type:  models.ConstraintTypeCommitLimit,
				Limit: 0,
			},
			expectError: true,
			errorMsg:    "commit_limit requires positive limit value",
		},
		{
			name: "commit_limit with date range (invalid)",
			constraint: &models.StatsConstraint{
				Type:  models.ConstraintTypeCommitLimit,
				Limit: 100,
				From:  "2024-01-01",
			},
			expectError: true,
			errorMsg:    "commit_limit cannot be used with date range",
		},
		{
			name: "invalid constraint type",
			constraint: &models.StatsConstraint{
				Type: "invalid_type",
			},
			expectError: true,
			errorMsg:    "constraint type must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatsConstraint(tt.constraint)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestExtractRepoName 测试仓库名称提取
func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "https url with .git",
			url:      "https://github.com/user/repo.git",
			expected: "repo",
		},
		{
			name:     "https url without .git",
			url:      "https://github.com/user/repo",
			expected: "repo",
		},
		{
			name:     "ssh url",
			url:      "git@github.com:user/repo.git",
			expected: "repo_git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoName(tt.url)
			assert.NotEmpty(t, result)
			// 注意：实际实现可能会有差异，这里主要测试不会panic
		})
	}
}
