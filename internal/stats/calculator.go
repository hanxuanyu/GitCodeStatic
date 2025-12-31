package stats

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/models"
)

// Calculator 统计计算器
type Calculator struct {
	gitPath string
}

// NewCalculator 创建统计计算器
func NewCalculator(gitPath string) *Calculator {
	if gitPath == "" {
		gitPath = "git"
	}
	return &Calculator{gitPath: gitPath}
}

// Calculate 计算统计数据
func (c *Calculator) Calculate(ctx context.Context, localPath, branch string, constraint *models.StatsConstraint) (*models.Statistics, error) {
	// 构建git log命令
	args := []string{
		"-C", localPath,
		"log",
		"--no-merges",
		"--numstat",
		"--pretty=format:COMMIT:%H|AUTHOR:%an|EMAIL:%ae|DATE:%ai",
	}

	// 添加约束条件
	if constraint != nil {
		if constraint.Type == models.ConstraintTypeDateRange {
			if constraint.From != "" {
				args = append(args, "--since="+constraint.From)
			}
			if constraint.To != "" {
				args = append(args, "--until="+constraint.To)
			}
		} else if constraint.Type == models.ConstraintTypeCommitLimit {
			args = append(args, "-n", strconv.Itoa(constraint.Limit))
		}
	}

	args = append(args, branch)

	logger.Logger.Debug().
		Str("local_path", localPath).
		Str("branch", branch).
		Interface("constraint", constraint).
		Msg("running git log")

	cmd := exec.CommandContext(ctx, c.gitPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git log: %w", err)
	}

	// 解析输出
	stats, err := c.parseGitLog(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse git log: %w", err)
	}

	// 填充摘要信息
	stats.Summary.TotalContributors = len(stats.ByContributor)
	if constraint != nil {
		if constraint.Type == models.ConstraintTypeDateRange {
			stats.Summary.DateRange = &models.DateRange{
				From: constraint.From,
				To:   constraint.To,
			}
		} else if constraint.Type == models.ConstraintTypeCommitLimit {
			stats.Summary.CommitLimit = &constraint.Limit
		}
	}

	return stats, nil
}

// parseGitLog 解析git log输出
func (c *Calculator) parseGitLog(output string) (*models.Statistics, error) {
	stats := &models.Statistics{
		Summary:       models.StatsSummary{},
		ByContributor: make([]models.ContributorStats, 0),
	}

	contributors := make(map[string]*models.ContributorStats)
	
	var currentAuthor, currentEmail string
	commitCount := 0

	scanner := bufio.NewScanner(strings.NewReader(output))
	commitPattern := regexp.MustCompile(`^COMMIT:(.+?)\|AUTHOR:(.+?)\|EMAIL:(.+?)\|DATE:(.+)$`)
	numstatPattern := regexp.MustCompile(`^(\d+|-)\s+(\d+|-)\s+(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" {
			continue
		}

		// 匹配提交行
		if matches := commitPattern.FindStringSubmatch(line); matches != nil {
			currentAuthor = matches[2]
			currentEmail = matches[3]
			commitCount++

			// 初始化贡献者统计
			if _, ok := contributors[currentEmail]; !ok {
				contributors[currentEmail] = &models.ContributorStats{
					Author: currentAuthor,
					Email:  currentEmail,
				}
			}
			contributors[currentEmail].Commits++
			continue
		}

		// 匹配文件变更行
		if matches := numstatPattern.FindStringSubmatch(line); matches != nil && currentEmail != "" {
			additionsStr := matches[1]
			deletionsStr := matches[2]

			// 处理二进制文件（显示为 -）
			additions := 0
			deletions := 0
			
			if additionsStr != "-" {
				additions, _ = strconv.Atoi(additionsStr)
			}
			if deletionsStr != "-" {
				deletions, _ = strconv.Atoi(deletionsStr)
			}

			contrib := contributors[currentEmail]
			contrib.Additions += additions
			contrib.Deletions += deletions
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading git log output: %w", err)
	}

	// 计算修改行数和净增加
	for _, contrib := range contributors {
		// 修改的定义：被替换的行数 = min(additions, deletions)
		contrib.Modifications = min(contrib.Additions, contrib.Deletions)
		contrib.NetAdditions = contrib.Additions - contrib.Deletions
		stats.ByContributor = append(stats.ByContributor, *contrib)
	}

	stats.Summary.TotalCommits = commitCount

	return stats, nil
}

// min 返回两个整数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
