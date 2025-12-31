package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/gitcodestatic/gitcodestatic/internal/models"
)

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(repoID int64, branch string, constraint *models.StatsConstraint, commitHash string) string {
	var constraintStr string
	
	if constraint != nil {
		if constraint.Type == models.ConstraintTypeDateRange {
			constraintStr = fmt.Sprintf("dr_%s_%s", constraint.From, constraint.To)
		} else if constraint.Type == models.ConstraintTypeCommitLimit {
			constraintStr = fmt.Sprintf("cl_%d", constraint.Limit)
		}
	}

	data := fmt.Sprintf("repo:%d|branch:%s|constraint:%s|commit:%s",
		repoID, branch, constraintStr, commitHash)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SerializeConstraint 序列化约束为JSON字符串
func SerializeConstraint(constraint *models.StatsConstraint) string {
	if constraint == nil {
		return "{}"
	}

	if constraint.Type == models.ConstraintTypeDateRange {
		return fmt.Sprintf(`{"type":"date_range","from":"%s","to":"%s"}`,
			constraint.From, constraint.To)
	} else if constraint.Type == models.ConstraintTypeCommitLimit {
		return fmt.Sprintf(`{"type":"commit_limit","limit":%d}`, constraint.Limit)
	}

	return "{}"
}
