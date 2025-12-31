package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/service"
	"github.com/go-chi/chi/v5"
)

// RepoHandler 仓库API处理器
type RepoHandler struct {
	repoService *service.RepoService
}

// NewRepoHandler 创建仓库处理器
func NewRepoHandler(repoService *service.RepoService) *RepoHandler {
	return &RepoHandler{
		repoService: repoService,
	}
}

// AddBatch 批量添加仓库
// @Summary 批量添加仓库
// @Description 批量添加多个Git仓库，异步克隆到本地
// @Tags 仓库管理
// @Accept json
// @Produce json
// @Param request body service.AddReposRequest true "仓库URL列表"
// @Success 200 {object} Response{data=service.AddReposResponse}
// @Failure 400 {object} Response
// @Router /repos/batch [post]
func (h *RepoHandler) AddBatch(w http.ResponseWriter, r *http.Request) {
	var req service.AddReposRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	if len(req.Repos) == 0 {
		respondError(w, http.StatusBadRequest, 40001, "repos cannot be empty")
		return
	}

	resp, err := h.repoService.AddRepos(r.Context(), &req)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to add repositories")
		respondError(w, http.StatusInternalServerError, 50000, "failed to add repositories")
		return
	}

	respondJSON(w, http.StatusOK, 0, "success", resp)
}

// List 获取仓库列表
// @Summary 获取仓库列表
// @Description 分页查询仓库列表，支持按状态筛选
// @Tags 仓库管理
// @Accept json
// @Produce json
// @Param status query string false "状态筛选(pending/cloning/ready/failed)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response
// @Router /repos [get]
func (h *RepoHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	repos, total, err := h.repoService.ListRepos(r.Context(), status, page, pageSize)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to list repositories")
		respondError(w, http.StatusInternalServerError, 50000, "failed to list repositories")
		return
	}

	data := map[string]interface{}{
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"repositories": repos,
	}

	respondJSON(w, http.StatusOK, 0, "success", data)
}

// Get 获取仓库详情
// @Summary 获取仓库详情
// @Description 根据ID获取仓库详细信息
// @Tags 仓库管理
// @Accept json
// @Produce json
// @Param id path int true "仓库ID"
// @Success 200 {object} Response{data=models.Repository}
// @Failure 404 {object} Response
// @Router /repos/{id} [get]
func (h *RepoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid repository id")
		return
	}

	repo, err := h.repoService.GetRepo(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, 40400, "repository not found")
		return
	}

	respondJSON(w, http.StatusOK, 0, "success", repo)
}

// SwitchBranch 切换分支
// @Summary 切换仓库分支
// @Description 异步切换仓库到指定分支
// @Tags 仓库管理
// @Accept json
// @Produce json
// @Param id path int true "仓库ID"
// @Param request body object{branch=string} true "分支名称"
// @Success 200 {object} Response{data=models.Task}
// @Failure 400 {object} Response
// @Router /repos/{id}/switch-branch [post]
func (h *RepoHandler) SwitchBranch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid repository id")
		return
	}

	var req struct {
		Branch string `json:"branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	if req.Branch == "" {
		respondError(w, http.StatusBadRequest, 40001, "branch cannot be empty")
		return
	}

	task, err := h.repoService.SwitchBranch(r.Context(), id, req.Branch)
	if err != nil {
		logger.Logger.Error().Err(err).Int64("repo_id", id).Msg("failed to switch branch")
		respondError(w, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, 0, "branch switch task submitted", task)
}

// Update 更新仓库
// @Summary 更新仓库
// @Description 异步拉取仓库最新代码(git pull)
// @Tags 仓库管理
// @Accept json
// @Produce json
// @Param id path int true "仓库ID"
// @Success 200 {object} Response{data=models.Task}
// @Failure 400 {object} Response
// @Router /repos/{id}/update [post]
func (h *RepoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid repository id")
		return
	}

	task, err := h.repoService.UpdateRepo(r.Context(), id)
	if err != nil {
		logger.Logger.Error().Err(err).Int64("repo_id", id).Msg("failed to update repository")
		respondError(w, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, 0, "update task submitted", task)
}

// Reset 重置仓库
// @Summary 重置仓库
// @Description 异步重置仓库到最新状态
// @Tags 仓库管理
// @Produce json
// @Param id path int true "仓库ID"
// @Success 200 {object} Response{data=models.Task}
// @Failure 400 {object} Response
// @Router /repos/{id}/reset [post]
func (h *RepoHandler) Reset(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid repository id")
		return
	}

	task, err := h.repoService.ResetRepo(r.Context(), id)
	if err != nil {
		logger.Logger.Error().Err(err).Int64("repo_id", id).Msg("failed to reset repository")
		respondError(w, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, 0, "reset task submitted", task)
}

// Delete 删除仓库
// @Summary 删除仓库
// @Description 删除指定仓库
// @Tags 仓库管理
// @Produce json
// @Param id path int true "仓库ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /repos/{id} [delete]
func (h *RepoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid repository id")
		return
	}

	if err := h.repoService.DeleteRepo(r.Context(), id); err != nil {
		logger.Logger.Error().Err(err).Int64("repo_id", id).Msg("failed to delete repository")
		respondError(w, http.StatusInternalServerError, 50000, "failed to delete repository")
		return
	}

	respondJSON(w, http.StatusOK, 0, "repository deleted successfully", nil)
}

// GetBranches 获取仓库分支列表
// @Summary 获取仓库分支列表
// @Description 获取指定仓库的所有分支
// @Tags 仓库管理
// @Produce json
// @Param id path int true "仓库ID"
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /repos/{id}/branches [get]
func (h *RepoHandler) GetBranches(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid repository id")
		return
	}

	branches, err := h.repoService.GetBranches(r.Context(), id)
	if err != nil {
		logger.Logger.Error().Err(err).Int64("repo_id", id).Msg("failed to get branches")
		respondError(w, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	data := map[string]interface{}{
		"branches": branches,
		"count":    len(branches),
	}

	respondJSON(w, http.StatusOK, 0, "success", data)
}
