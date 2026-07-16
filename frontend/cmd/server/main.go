// Local development server. Manually wires the same routes that Vercel exposes through individual functions.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"dimoo-tracker-frontend/internal/app"
)

func main() {
	loadDotEnv()

	addr := ":3000"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Printf("frontend listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, app.New()))
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || os.Getenv(strings.TrimSpace(key)) != "" {
			continue
		}
		os.Setenv(strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), `"'`))
	}
}
