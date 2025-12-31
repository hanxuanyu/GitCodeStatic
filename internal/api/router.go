package api

import (
	"net/http"

	_ "github.com/gitcodestatic/gitcodestatic/docs"
	"github.com/gitcodestatic/gitcodestatic/internal/api/handlers"
	"github.com/gitcodestatic/gitcodestatic/internal/service"
	"github.com/gitcodestatic/gitcodestatic/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Router 路由配置
type Router struct {
	repoHandler  *handlers.RepoHandler
	statsHandler *handlers.StatsHandler
	taskHandler  *handlers.TaskHandler
	webDir       string
	webEnabled   bool
}

// NewRouter 创建路由
func NewRouter(repoService *service.RepoService, statsService *service.StatsService, store storage.Store, webDir string, webEnabled bool) *Router {
	return &Router{
		repoHandler:  handlers.NewRepoHandler(repoService),
		statsHandler: handlers.NewStatsHandler(statsService, store),
		taskHandler:  handlers.NewTaskHandler(store),
		webDir:       webDir,
		webEnabled:   webEnabled,
	}
}

// Setup 设置路由
func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Web UI static files
	if rt.webEnabled {
		fileServer := http.FileServer(http.Dir(rt.webDir))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			fileServer.ServeHTTP(w, r)
		})
	}

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// 仓库管理
		r.Route("/repos", func(r chi.Router) {
			r.Post("/batch", rt.repoHandler.AddBatch)
			r.Get("/", rt.repoHandler.List)
			r.Get("/{id}", rt.repoHandler.Get)
			r.Get("/{id}/branches", rt.repoHandler.GetBranches)
			r.Post("/{id}/switch-branch", rt.repoHandler.SwitchBranch)
			r.Post("/{id}/update", rt.repoHandler.Update)
			r.Post("/{id}/reset", rt.repoHandler.Reset)
			r.Delete("/{id}", rt.repoHandler.Delete)
		})

		// 统计
		r.Route("/stats", func(r chi.Router) {
			r.Post("/calculate", rt.statsHandler.Calculate)
			r.Get("/result", rt.statsHandler.QueryResult)
			r.Get("/commit-count", rt.statsHandler.CountCommits)
			r.Get("/caches", rt.statsHandler.ListCaches)
			r.Delete("/caches/clear", rt.statsHandler.ClearAllCaches)
		})

		// 任务
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", rt.taskHandler.List)
			r.Delete("/clear", rt.taskHandler.ClearAllTasks)
			r.Delete("/clear-completed", rt.taskHandler.ClearCompletedTasks)
		})
	})

	return r
}
