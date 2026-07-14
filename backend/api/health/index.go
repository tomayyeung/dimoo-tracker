package handler

import (
	"net/http"

	"dimoo-tracker-backend/internal/httpx"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if httpx.WithCORS(w, r) {
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
