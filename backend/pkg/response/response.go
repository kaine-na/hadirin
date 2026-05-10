package response

import (
	"encoding/json"
	"net/http"
)

// Response adalah format standar semua API response.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedData membungkus data list dengan metadata pagination.
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success mengirim response sukses dengan status 200.
func Success(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created mengirim response sukses dengan status 201.
func Created(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error mengirim response error dengan status code yang diberikan.
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, Response{
		Success: false,
		Message: message,
	})
}

// Paginated mengirim response list dengan metadata pagination.
func Paginated(w http.ResponseWriter, message string, items interface{}, total, page, pageSize int) {
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	JSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Data: PaginatedData{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// JSON menulis JSON response dengan header yang benar.
func JSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
