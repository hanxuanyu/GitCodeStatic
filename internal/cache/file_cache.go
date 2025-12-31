package cache

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hanxuanyu/gitcodestatic/internal/logger"
	"github.com/hanxuanyu/gitcodestatic/internal/models"
	"github.com/hanxuanyu/gitcodestatic/internal/storage"
)

// FileCache 基于文件+DB的缓存实现
type FileCache struct {
	store    storage.Store
	statsDir string
}

// NewFileCache 创建文件缓存
func NewFileCache(store storage.Store, statsDir string) *FileCache {
	return &FileCache{
		store:    store,
		statsDir: statsDir,
	}
}

// Get 获取缓存
func (c *FileCache) Get(ctx context.Context, cacheKey string) (*models.StatsResult, error) {
	// 从DB查询缓存元数据
	cache, err := c.store.StatsCache().GetByCacheKey(ctx, cacheKey)
	if err != nil {
		return nil, err
	}
	if cache == nil {
		return nil, nil // 缓存不存在
	}

	// 读取结果文件
	stats, err := c.loadStatsFromFile(cache.ResultPath)
	if err != nil {
		logger.Logger.Error().Err(err).Str("cache_key", cacheKey).Msg("failed to load stats from file")
		return nil, err
	}

	// 更新命中次数
	if err := c.store.StatsCache().UpdateHitCount(ctx, cache.ID); err != nil {
		logger.Logger.Warn().Err(err).Int64("cache_id", cache.ID).Msg("failed to update hit count")
	}

	result := &models.StatsResult{
		CacheHit:   true,
		CachedAt:   &cache.CreatedAt,
		CommitHash: cache.CommitHash,
		Statistics: stats,
	}

	logger.Logger.Info().
		Str("cache_key", cacheKey).
		Int64("cache_id", cache.ID).
		Msg("cache hit")

	return result, nil
}

// Set 设置缓存
func (c *FileCache) Set(ctx context.Context, repoID int64, branch string, constraint *models.StatsConstraint,
	commitHash string, stats *models.Statistics) error {

	// 生成缓存键
	cacheKey := GenerateCacheKey(repoID, branch, constraint, commitHash)

	// 保存统计结果到文件
	resultPath := filepath.Join(c.statsDir, cacheKey+".json.gz")
	if err := c.saveStatsToFile(stats, resultPath); err != nil {
		return fmt.Errorf("failed to save stats to file: %w", err)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(resultPath)
	if err != nil {
		return fmt.Errorf("failed to stat result file: %w", err)
	}

	// 创建缓存记录
	cache := &models.StatsCache{
		RepoID:          repoID,
		Branch:          branch,
		ConstraintType:  constraint.Type,
		ConstraintValue: SerializeConstraint(constraint),
		CommitHash:      commitHash,
		ResultPath:      resultPath,
		ResultSize:      fileInfo.Size(),
		CacheKey:        cacheKey,
	}

	if err := c.store.StatsCache().Create(ctx, cache); err != nil {
		// 如果创建失败，删除已保存的文件
		os.Remove(resultPath)
		return fmt.Errorf("failed to create cache record: %w", err)
	}

	logger.Logger.Info().
		Str("cache_key", cacheKey).
		Int64("cache_id", cache.ID).
		Int64("file_size", fileInfo.Size()).
		Msg("cache saved")

	return nil
}

// InvalidateByRepoID 使指定仓库的所有缓存失效
func (c *FileCache) InvalidateByRepoID(ctx context.Context, repoID int64) error {
	// 查询该仓库的所有缓存
	// 注意：这里简化实现，实际应该先查询再删除文件
	if err := c.store.StatsCache().DeleteByRepoID(ctx, repoID); err != nil {
		return fmt.Errorf("failed to delete cache records: %w", err)
	}

	logger.Logger.Info().Int64("repo_id", repoID).Msg("cache invalidated")
	return nil
}

// saveStatsToFile 保存统计结果到文件（gzip压缩）
func (c *FileCache) saveStatsToFile(stats *models.Statistics, filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 创建gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// 编码JSON
	encoder := json.NewEncoder(gzipWriter)
	if err := encoder.Encode(stats); err != nil {
		return fmt.Errorf("failed to encode stats: %w", err)
	}

	return nil
}

// loadStatsFromFile 从文件加载统计结果
func (c *FileCache) loadStatsFromFile(filePath string) (*models.Statistics, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 创建gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// 解码JSON
	var stats models.Statistics
	decoder := json.NewDecoder(gzipReader)
	if err := decoder.Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return &stats, nil
}
