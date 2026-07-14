package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPagesRender(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/series":
			_, _ = w.Write([]byte(`[{"id":"series-1","name":"Dimoo Dream Travel","theme":"clouds","release_year":2024}]`))
		case "/api/figurines", "/api/collection", "/api/wishlist", "/api/shelf":
			_, _ = w.Write([]byte(`[{"id":"fig-1","series_id":"series-1","series_name":"Dimoo Dream Travel","name":"Cloud Boarding Pass","character":"Dimoo","rarity":"standard","owned":true,"wishlisted":false,"on_shelf":true}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	t.Setenv("API_BASE_URL", backend.URL)
	app := New()

	for _, path := range []string{"/", "/collection", "/search", "/wishlist", "/static/styles.css"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, rec.Code, rec.Body.String())
		}
	}
}

func TestActionRejectsExternalRedirect(t *testing.T) {
	os.Setenv("API_BASE_URL", "http://127.0.0.1:1")
	t.Cleanup(func() { os.Unsetenv("API_BASE_URL") })

	req := httptest.NewRequest(http.MethodPost, "/actions/collection/add", strings.NewReader("figurine_id=fig-1&next=https://example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)
	if rec.Code == http.StatusSeeOther && rec.Header().Get("Location") == "https://example.com" {
		t.Fatal("external redirect was allowed")
	}
}
