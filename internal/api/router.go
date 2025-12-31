package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gitcodestatic/gitcodestatic/internal/api/handlers"
	"github.com/gitcodestatic/gitcodestatic/internal/service"
)

// Router 路由配置
type Router struct {
	repoHandler  *handlers.RepoHandler
	statsHandler *handlers.StatsHandler
}

// NewRouter 创建路由
func NewRouter(repoService *service.RepoService, statsService *service.StatsService) *Router {
	return &Router{
		repoHandler:  handlers.NewRepoHandler(repoService),
		statsHandler: handlers.NewStatsHandler(statsService),
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

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// 仓库管理
		r.Route("/repos", func(r chi.Router) {
			r.Post("/batch", rt.repoHandler.AddBatch)
			r.Get("/", rt.repoHandler.List)
			r.Get("/{id}", rt.repoHandler.Get)
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
		})
	})

	return r
}
