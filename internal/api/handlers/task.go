package handlers

import (
	"net/http"
	"strconv"

	"github.com/gitcodestatic/gitcodestatic/internal/logger"
	"github.com/gitcodestatic/gitcodestatic/internal/storage"
)

// TaskHandler 任务API处理器
type TaskHandler struct {
	store storage.Store
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(store storage.Store) *TaskHandler {
	return &TaskHandler{
		store: store,
	}
}

// List 查询任务列表
// @Summary 查询任务列表
// @Description 查询任务列表，可按状态过滤
// @Tags 任务管理
// @Produce json
// @Param status query string false "任务状态"
// @Param limit query int false "返回数量限制" default(50)
// @Success 200 {object} Response{data=object}
// @Failure 500 {object} Response
// @Router /tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
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

	// 使用 List 方法，repoID=0 表示不过滤仓库
	tasks, total, err := h.store.Tasks().List(r.Context(), 0, status, 1, limit)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to list tasks")
		respondError(w, http.StatusInternalServerError, 50000, "failed to list tasks")
		return
	}

	data := map[string]interface{}{
		"tasks": tasks,
		"total": total,
	}

	respondJSON(w, http.StatusOK, 0, "success", data)
}

// ClearAllTasks 清除所有任务记录
// @Summary 清除所有任务记录
// @Description 删除所有任务记录（包括进行中的）
// @Tags 任务管理
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /tasks/clear [delete]
func (h *TaskHandler) ClearAllTasks(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Tasks().DeleteAll(r.Context()); err != nil {
		logger.Logger.Error().Err(err).Msg("failed to clear all tasks")
		respondError(w, http.StatusInternalServerError, 50000, "failed to clear tasks")
		return
	}

	logger.Logger.Info().Msg("all tasks cleared")
	respondJSON(w, http.StatusOK, 0, "所有任务记录已清除", nil)
}

// ClearCompletedTasks 清除已完成的任务记录
// @Summary 清除已完成的任务记录
// @Description 删除已完成、失败或取消的任务记录
// @Tags 任务管理
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /tasks/clear-completed [delete]
func (h *TaskHandler) ClearCompletedTasks(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Tasks().DeleteCompleted(r.Context()); err != nil {
		logger.Logger.Error().Err(err).Msg("failed to clear completed tasks")
		respondError(w, http.StatusInternalServerError, 50000, "failed to clear completed tasks")
		return
	}

	logger.Logger.Info().Msg("completed tasks cleared")
	respondJSON(w, http.StatusOK, 0, "已完成的任务记录已清除", nil)
}
