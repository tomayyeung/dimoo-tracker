// Package handler exposes the Vercel function for catalog series routes.
package handler

import (
	"net/http"

	"dimoo-tracker-backend/internal/db"
	"dimoo-tracker-backend/internal/httpx"
)

// Handler serves GET /api/series.
func Handler(w http.ResponseWriter, r *http.Request) {
	if httpx.WithCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	items, err := db.Series(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}
