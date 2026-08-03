// Package handler exposes the Vercel function for backend health checks.
package handler

import (
	"net/http"

	"dimoo-tracker-backend/internal/httpx"
)

// Handler serves /api/health with a simple JSON status response.
func Handler(w http.ResponseWriter, r *http.Request) {
	if httpx.WithCORS(w, r) {
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
