package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Reason string `json:"reason"`
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Reason: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding error response", http.StatusInternalServerError)
	}
}
