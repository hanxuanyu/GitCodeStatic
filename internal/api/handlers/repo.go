package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/service"
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
func (h *RepoHandler) AddBatch(w http.ResponseWriter, r *http.Request) {
	var req service.AddReposRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	if len(req.URLs) == 0 {
		respondError(w, http.StatusBadRequest, 40001, "urls cannot be empty")
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
