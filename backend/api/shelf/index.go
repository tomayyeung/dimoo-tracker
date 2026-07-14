package handler

import (
	"encoding/json"
	"net/http"

	"dimoo-tracker-backend/internal/db"
	"dimoo-tracker-backend/internal/httpx"
	"dimoo-tracker-backend/internal/models"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if httpx.WithCORS(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := db.Shelf(r.Context())
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.JSON(w, http.StatusOK, items)
	case http.MethodPost:
		id, ok := figurineID(w, r)
		if !ok {
			return
		}
		if err := db.AddShelf(r.Context(), id); err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			httpx.Error(w, http.StatusBadRequest, "id is required")
			return
		}
		if err := db.RemoveShelf(r.Context(), id); err != nil {
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func figurineID(w http.ResponseWriter, r *http.Request) (string, bool) {
	var input models.FigurineInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.FigurineID == "" {
		httpx.Error(w, http.StatusBadRequest, "figurine_id is required")
		return "", false
	}
	return input.FigurineID, true
}
