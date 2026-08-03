// Package httpx contains small HTTP response helpers shared by backend API
// functions.
package httpx

import (
	"encoding/json"
	"net/http"
	"os"
)

// WithCORS writes CORS headers and handles OPTIONS preflight requests.
//
// It returns true when the request has already been answered and the caller
// should stop processing.
func WithCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

// JSON writes v as an application/json response with status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes a JSON error response in the shape expected by the frontend.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
