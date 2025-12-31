package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/service"
)

// StatsHandler 统计API处理器
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// Calculate 触发统计计算
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
