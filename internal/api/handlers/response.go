package handlers

import (
	"encoding/json"
	"net/http"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// respondJSON 返回JSON响应
func respondJSON(w http.ResponseWriter, statusCode, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	
	json.NewEncoder(w).Encode(resp)
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, statusCode, code int, message string) {
	respondJSON(w, statusCode, code, message, nil)
}
