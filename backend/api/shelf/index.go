// Package handler exposes the Vercel function for shelf item routes.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"dimoo-tracker-backend/internal/db"
	"dimoo-tracker-backend/internal/httpx"
	"dimoo-tracker-backend/internal/models"
)

// Handler serves /api/shelf requests.
//
// GET returns featured shelf figurines, POST accepts {"figurine_id":"..."},
// PATCH swaps two shelf items, and DELETE removes the item identified by the id query parameter.
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
			if errors.Is(err, db.ErrNotOwned) {
				httpx.Error(w, http.StatusConflict, err.Error())
				return
			}
			httpx.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
	case http.MethodPatch:
		input, ok := shelfSwapInput(w, r)
		if !ok {
			return
		}
		if input.FigurineID == input.TargetFigurineID {
			httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
			return
		}
		if err := db.SwapShelf(r.Context(), input.FigurineID, input.TargetFigurineID); err != nil {
			if errors.Is(err, db.ErrNotShelf) {
				httpx.Error(w, http.StatusConflict, err.Error())
				return
			}
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

func shelfSwapInput(w http.ResponseWriter, r *http.Request) (models.ShelfSwapInput, bool) {
	var input models.ShelfSwapInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.FigurineID == "" || input.TargetFigurineID == "" {
		httpx.Error(w, http.StatusBadRequest, "figurine_id and target_figurine_id are required")
		return models.ShelfSwapInput{}, false
	}
	return input, true
}
