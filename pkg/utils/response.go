package utils

import (
	"encoding/json"
	"net/http"
)

// WriteSuccessResponse writes a success response
func WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, message, http.StatusBadRequest)
}

// WriteUnauthorizedError writes an unauthorized error response
func WriteUnauthorizedError(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, message, http.StatusUnauthorized)
}

// WriteNotFoundError writes a not found error response
func WriteNotFoundError(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, message, http.StatusNotFound)
}

// WriteInternalServerError writes an internal server error response
func WriteInternalServerError(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, message, http.StatusInternalServerError)
}
