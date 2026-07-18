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
			_, _ = w.Write([]byte(`[{"id":"series-1","name":"Dimoo Dream Travel","ip":"Dimoo","release_year":2024}]`))
		case "/api/figurines", "/api/collection", "/api/wishlist", "/api/shelf":
			_, _ = w.Write([]byte(`[{"id":"fig-1","series_id":"series-1","series_name":"Dimoo Dream Travel","name":"Cloud Boarding Pass","rarity":"standard","owned":true,"wishlisted":false,"on_shelf":true}]`))
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

func TestFigurineCardActionsUseHTMXPageSwap(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/series":
			_, _ = w.Write([]byte(`[{"id":"series-1","name":"Dimoo Dream Travel","ip":"Dimoo","release_year":2024}]`))
		case "/api/figurines":
			_, _ = w.Write([]byte(`[{"id":"fig-1","series_id":"series-1","series_name":"Dimoo Dream Travel","name":"Cloud Boarding Pass","rarity":"standard","owned":true,"wishlisted":false,"on_shelf":true}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	t.Setenv("API_BASE_URL", backend.URL)
	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search returned %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`hx-post="/actions/collection/remove"`,
		`hx-target="closest .figurine-card"`,
		`hx-select=".figurine-card"`,
		`hx-swap="outerHTML show:none focus-scroll:false"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered card missing %s", want)
		}
	}
	if strings.Contains(body, `button class="button ghost" type="submit" hx-post=`) {
		t.Fatal("card action button should not own the HTMX post")
	}
}

func TestSearchSendsIPFilter(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/series":
			_, _ = w.Write([]byte(`[{"id":"series-1","name":"Dimoo Dream Travel","ip":"Dimoo","release_year":2024}]`))
		case "/api/figurines":
			if got := r.URL.Query().Get("ip"); got != "Dimoo" {
				t.Fatalf("figurines request ip = %q, want Dimoo", got)
			}
			if got := r.URL.Query().Get("character"); got != "" {
				t.Fatalf("figurines request still sent character = %q", got)
			}
			_, _ = w.Write([]byte(`[]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	t.Setenv("API_BASE_URL", backend.URL)
	req := httptest.NewRequest(http.MethodGet, "/search?ip=Dimoo", nil)
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search returned %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHTMXSearchActionReturnsFigurineCardFragment(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/collection" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/api/figurines" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`[{"id":"fig-1","series_id":"series-1","series_name":"Dimoo Dream Travel","name":"Cloud Boarding Pass","rarity":"standard","owned":true,"wishlisted":false,"on_shelf":false}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	t.Setenv("API_BASE_URL", backend.URL)
	req := httptest.NewRequest(http.MethodPost, "/actions/collection/add", strings.NewReader("figurine_id=fig-1&next=/search?q=Cloud"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("action returned %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `<article class="figurine-card`) {
		t.Fatal("HTMX search action response should include a card fragment")
	}
	if strings.Contains(body, `<main class="page-shell">`) || strings.Contains(body, "<!doctype html>") {
		t.Fatal("HTMX search action response should not include the page shell or full document")
	}
}

func TestHTMXActionReturnsPageShellFragment(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/collection" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/api/collection" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`[{"id":"fig-1","series_id":"series-1","series_name":"Dimoo Dream Travel","name":"Cloud Boarding Pass","rarity":"standard","owned":true,"wishlisted":false,"on_shelf":false}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	t.Setenv("API_BASE_URL", backend.URL)
	req := httptest.NewRequest(http.MethodPost, "/actions/collection/add", strings.NewReader("figurine_id=fig-1&next=/collection"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("action returned %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Header().Get("HX-Redirect") != "" {
		t.Fatal("HTMX action should not force a full-page HX-Redirect")
	}
	if rec.Header().Get("Location") != "" {
		t.Fatalf("HTMX action should not return Location header, got %q", rec.Header().Get("Location"))
	}
	body := rec.Body.String()
	if !strings.Contains(body, `<main class="page-shell">`) {
		t.Fatal("HTMX action response should include the page shell fragment")
	}
	if strings.Contains(body, "<!doctype html>") {
		t.Fatal("HTMX action response should not include the full document")
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
