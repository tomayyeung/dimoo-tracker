package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	collection "dimoo-tracker-backend/api/collection"
	figurines "dimoo-tracker-backend/api/figurines"
	health "dimoo-tracker-backend/api/health"
	series "dimoo-tracker-backend/api/series"
	shelf "dimoo-tracker-backend/api/shelf"
	wishlist "dimoo-tracker-backend/api/wishlist"
)

func main() {
	loadDotEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", health.Handler)
	mux.HandleFunc("/api/series", series.Handler)
	mux.HandleFunc("/api/figurines", figurines.Handler)
	mux.HandleFunc("/api/collection", collection.Handler)
	mux.HandleFunc("/api/wishlist", wishlist.Handler)
	mux.HandleFunc("/api/shelf", shelf.Handler)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Printf("backend listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
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
