package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/service"
	"github.com/gitcodestatic/gitcodestatic/internal/storage"
)

// StatsHandler 统计API处理器
type StatsHandler struct {
	statsService *service.StatsService
	store        storage.Store
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(statsService *service.StatsService, store storage.Store) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		store:        store,
	}
}

// Calculate 触发统计计算
// @Summary 触发统计任务
// @Description 异步触发统计计算任务
// @Tags 统计管理
// @Accept json
// @Produce json
// @Param request body service.CalculateRequest true "统计请求"
// @Success 200 {object} Response{data=models.Task}
// @Failure 400 {object} Response
// @Router /stats/calculate [post]
func (h *StatsHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req service.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	if req.RepoID == 0 {
		respondError(w, http.StatusBadRequest, 40001, "repo_id is required")
		return
	}

	if req.Branch == "" {
		respondError(w, http.StatusBadRequest, 40001, "branch is required")
		return
	}

	// 校验约束参数
	if err := service.ValidateStatsConstraint(req.Constraint); err != nil {
		respondError(w, http.StatusBadRequest, 40001, err.Error())
		return
	}

	task, err := h.statsService.Calculate(r.Context(), &req)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to submit stats task")
		respondError(w, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, 0, "statistics task submitted", task)
}

// QueryResult 查询统计结果
// @Summary 查询统计结果
// @Description 查询统计计算结果
// @Tags 统计管理
// @Produce json
// @Param repo_id query int true "仓库ID"
// @Param branch query string true "分支名称"
// @Param constraint_type query string false "约束类型"
// @Param from query string false "开始日期"
// @Param to query string false "结束日期"
// @Param limit query int false "提交数限制"
// @Success 200 {object} Response{data=models.StatsResult}
// @Failure 400 {object} Response
// @Router /stats/query [get]
func (h *StatsHandler) QueryResult(w http.ResponseWriter, r *http.Request) {
	repoID, _ := strconv.ParseInt(r.URL.Query().Get("repo_id"), 10, 64)
	branch := r.URL.Query().Get("branch")
	constraintType := r.URL.Query().Get("constraint_type")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if repoID == 0 {
		respondError(w, http.StatusBadRequest, 40001, "repo_id is required")
		return
	}

	if branch == "" {
		respondError(w, http.StatusBadRequest, 40001, "branch is required")
		return
	}

	req := &service.QueryResultRequest{
		RepoID:         repoID,
		Branch:         branch,
		ConstraintType: constraintType,
		From:           from,
		To:             to,
		Limit:          limit,
	}

	result, err := h.statsService.QueryResult(r.Context(), req)
	if err != nil {
		if err.Error() == "statistics not found, please submit calculation task first" {
			respondError(w, http.StatusNotFound, 40400, err.Error())
			return
		}
		logger.Logger.Error().Err(err).Msg("failed to query stats result")
		respondError(w, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, 0, "success", result)
}

// CountCommits 统计提交次数
// @Summary 统计提交次数
// @Description 统计指定条件下的提交次数
// @Tags 统计管理
// @Produce json
// @Param repo_id query int true "仓库ID"
// @Param branch query string true "分支名称"
// @Param from query string false "开始日期"
// @Success 200 {object} Response{data=service.CountCommitsResponse}
// @Failure 400 {object} Response
// @Router /stats/commits/count [get]
func (h *StatsHandler) CountCommits(w http.ResponseWriter, r *http.Request) {
	repoID, _ := strconv.ParseInt(r.URL.Query().Get("repo_id"), 10, 64)
	branch := r.URL.Query().Get("branch")
	from := r.URL.Query().Get("from")

	if repoID == 0 {
		respondError(w, http.StatusBadRequest, 40001, "repo_id is required")
		return
	}

	if branch == "" {
		respondError(w, http.StatusBadRequest, 40001, "branch is required")
		return
	}

	req := &service.CountCommitsRequest{
		RepoID: repoID,
		Branch: branch,
		From:   from,
	}

	result, err := h.statsService.CountCommits(r.Context(), req)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to count commits")
		respondError(w, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, 0, "success", result)
}

// ListCaches 获取统计缓存列表
// @Summary 获取统计缓存列表
// @Description 获取已计算的统计缓存列表
// @Tags 统计管理
// @Produce json
// @Param repo_id query int false "仓库ID（可选，不传则返回所有）"
// @Param limit query int false "返回数量限制" default(50)
// @Success 200 {object} Response{data=object}
// @Failure 500 {object} Response
// @Router /stats/caches [get]
func (h *StatsHandler) ListCaches(w http.ResponseWriter, r *http.Request) {
	repoID, _ := strconv.ParseInt(r.URL.Query().Get("repo_id"), 10, 64)
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 200 {
				limit = 200
			}
		}
	}

	caches, total, err := h.store.StatsCache().List(r.Context(), repoID, limit)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to list stats caches")
		respondError(w, http.StatusInternalServerError, 50000, "failed to list stats caches")
		return
	}

	data := map[string]interface{}{
		"caches": caches,
		"total":  total,
	}

	respondJSON(w, http.StatusOK, 0, "success", data)
}

// ClearAllCaches 清除所有统计缓存
// @Summary 清除所有统计缓存
// @Description 删除所有统计缓存记录和文件
// @Tags 统计管理
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /stats/caches/clear [delete]
func (h *StatsHandler) ClearAllCaches(w http.ResponseWriter, r *http.Request) {
	// 删除数据库中的缓存记录
	if err := h.store.StatsCache().DeleteAll(r.Context()); err != nil {
		logger.Logger.Error().Err(err).Msg("failed to clear all caches")
		respondError(w, http.StatusInternalServerError, 50000, "failed to clear caches")
		return
	}

	logger.Logger.Info().Msg("all stats caches cleared")
	respondJSON(w, http.StatusOK, 0, "所有统计缓存已清除", nil)
}
