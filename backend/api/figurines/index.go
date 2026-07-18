package handler

import (
	"net/http"

	"dimoo-tracker-backend/internal/db"
	"dimoo-tracker-backend/internal/httpx"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if httpx.WithCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	query := r.URL.Query()
	items, err := db.Figurines(r.Context(), query.Get("q"), query.Get("series_id"), query.Get("ip"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}
