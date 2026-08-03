// Package handler exposes the Vercel function for the frontend application.
package handler

import (
	"net/http"

	"dimoo-tracker-frontend/internal/app"
)

// Handler serves all frontend routes through the Go-rendered HTMX app.
func Handler(w http.ResponseWriter, r *http.Request) {
	app.New().ServeHTTP(w, r)
}
