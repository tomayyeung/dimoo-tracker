package handler

import (
	"net/http"

	"dimoo-tracker-frontend/internal/app"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	app.New().ServeHTTP(w, r)
}
