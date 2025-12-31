package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/gitcodestatic/gitcodestatic/docs"
	"github.com/gitcodestatic/gitcodestatic/internal/api"
	"github.com/gitcodestatic/gitcodestatic/internal/cache"
	"github.com/gitcodestatic/gitcodestatic/internal/config"
	"github.com/gitcodestatic/gitcodestatic/internal/git"
	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/models"
	"github.com/gitcodestatic/gitcodestatic/internal/service"
	"github.com/gitcodestatic/gitcodestatic/internal/stats"
	"github.com/gitcodestatic/gitcodestatic/internal/storage/sqlite"
	"github.com/gitcodestatic/gitcodestatic/internal/worker"
)

func main() {
	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.InitLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Logger.Info().Msg("starting GitCodeStatic server")

	// 创建工作目录
	if err := ensureDirectories(cfg); err != nil {
		logger.Logger.Fatal().Err(err).Msg("failed to create directories")
	}

	// 初始化存储
	store, err := sqlite.NewSQLiteStore(cfg.Storage.SQLite.Path)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("failed to create store")
	}
	defer store.Close()

	if err := store.Init(); err != nil {
		logger.Logger.Fatal().Err(err).Msg("failed to initialize database")
	}

	logger.Logger.Info().Msg("database initialized")

	// 创建Git管理器
	gitManager := git.NewCmdGitManager(cfg.Git.CommandPath)
	if !gitManager.IsAvailable() {
		logger.Logger.Warn().Msg("git command not available, some features may not work")
	} else {
		logger.Logger.Info().Msg("git command available")
	}

	// 创建统计计算器
	calculator := stats.NewCalculator(cfg.Git.CommandPath)

	// 创建缓存
	fileCache := cache.NewFileCache(store, cfg.Workspace.StatsDir)

	// 创建任务队列
	queue := worker.NewQueue(cfg.Worker.QueueBuffer, store)

	// 创建任务处理器
	handlers := map[string]worker.TaskHandler{
		models.TaskTypeClone:  worker.NewCloneHandler(store, gitManager),
		models.TaskTypePull:   worker.NewPullHandler(store, gitManager),
		models.TaskTypeSwitch: worker.NewSwitchHandler(store, gitManager),
		models.TaskTypeReset:  worker.NewResetHandler(store, gitManager, fileCache),
		models.TaskTypeStats:  worker.NewStatsHandler(store, calculator, fileCache, gitManager),
	}

	// 创建Worker池
	totalWorkers := cfg.Worker.CloneWorkers + cfg.Worker.PullWorkers +
		cfg.Worker.StatsWorkers + cfg.Worker.GeneralWorkers

	pool := worker.NewPool(totalWorkers, queue, store, handlers)
	pool.Start()
	defer pool.Stop()

	logger.Logger.Info().Int("workers", totalWorkers).Msg("worker pool started")

	// 创建服务层
	repoService := service.NewRepoService(store, queue, cfg.Workspace.CacheDir, gitManager)
	statsService := service.NewStatsService(store, queue, fileCache, gitManager)

	// 设置路由
	router := api.NewRouter(repoService, statsService, store, cfg.Web.Dir, cfg.Web.Enabled)
	handler := router.Setup()

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	go func() {
		logger.Logger.Info().Str("addr", addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info().Msg("shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error().Err(err).Msg("server forced to shutdown")
	}

	logger.Logger.Info().Msg("server stopped")
}

// loadConfig 加载配置
func loadConfig() (*config.Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Logger.Warn().Str("path", configPath).Msg("config file not found, using defaults")
		return config.DefaultConfig(), nil
	}

	return config.LoadConfig(configPath)
}

// ensureDirectories 确保工作目录存在
func ensureDirectories(cfg *config.Config) error {
	dirs := []string{
		cfg.Workspace.BaseDir,
		cfg.Workspace.CacheDir,
		cfg.Workspace.StatsDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
