package app

import (
	"html/template"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"

	"dimoo-tracker-frontend/assets"
	"dimoo-tracker-frontend/internal/backend"
)

type App struct {
	backend backend.Client
}

type PageData struct {
	Title          string
	Active         string // Selected nav tab
	Error          string // Displays top-page error banner
	Figurines      []backend.Figurine
	Series         []backend.Series // Search filter options
	IPs            []string         // Character/collaboration filter options
	SelectedQuery  string
	SelectedSeries string
	SelectedIP     string
	Next           string // Where action handlers redirect after add/remove
}

// Create app with backend client
func New() App {
	return App{backend: backend.New()}
}

// Main frontend router.
func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		a.static(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/actions/") {
		a.action(w, r)
		return
	}
	if r.URL.Path == "/partials/search" {
		a.searchPartial(w, r)
		return
	}

	switch r.URL.Path {
	case "/":
		a.shelf(w, r)
	case "/collection":
		a.collection(w, r)
	case "/search":
		a.search(w, r)
	case "/wishlist":
		a.wishlist(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a App) shelf(w http.ResponseWriter, r *http.Request) {
	items, err := a.backend.Shelf(r.Context())
	data := PageData{Title: "Shelf", Active: "shelf", Figurines: items, Next: "/"}
	if err != nil {
		data.Error = err.Error()
	}
	a.render(w, "shelf.html", data)
}

func (a App) collection(w http.ResponseWriter, r *http.Request) {
	items, err := a.backend.Collection(r.Context())
	data := PageData{Title: "My Collection", Active: "collection", Figurines: items, Next: "/collection"}
	if err != nil {
		data.Error = err.Error()
	}
	a.render(w, "collection.html", data)
}

func (a App) search(w http.ResponseWriter, r *http.Request) {
	data := a.searchData(w, r)
	a.render(w, "search.html", data)
}

func (a App) searchPartial(w http.ResponseWriter, r *http.Request) {
	data := a.searchData(w, r)
	t := a.templates("search.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.ExecuteTemplate(w, "search_results", data)
}

func (a App) searchData(_ http.ResponseWriter, r *http.Request) PageData {
	query := r.URL.Query()
	items, err := a.backend.Figurines(r.Context(), query.Get("q"), query.Get("series_id"), query.Get("ip"))
	series, seriesErr := a.backend.Series(r.Context())
	next := r.URL.RequestURI()
	if strings.HasPrefix(next, "/partials/search") {
		next = "/search"
		if r.URL.RawQuery != "" {
			next += "?" + r.URL.RawQuery
		}
	}
	data := PageData{
		Title:          "Search",
		Active:         "search",
		Figurines:      items,
		Series:         series,
		IPs:            ips(series),
		SelectedQuery:  query.Get("q"),
		SelectedSeries: query.Get("series_id"),
		SelectedIP:     query.Get("ip"),
		Next:           next,
	}
	if err != nil {
		data.Error = err.Error()
	} else if seriesErr != nil {
		data.Error = seriesErr.Error()
	}
	return data
}

func (a App) wishlist(w http.ResponseWriter, r *http.Request) {
	items, err := a.backend.Wishlist(r.Context())
	data := PageData{Title: "Wishlist", Active: "wishlist", Figurines: items, Next: "/wishlist"}
	if err != nil {
		data.Error = err.Error()
	}
	a.render(w, "wishlist.html", data)
}

func (a App) action(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := r.FormValue("figurine_id")
	next := r.FormValue("next")
	if next == "" || !strings.HasPrefix(next, "/") {
		next = "/"
	}
	var err error
	switch r.URL.Path {
	case "/actions/collection/add":
		err = a.backend.AddCollection(r.Context(), id)
	case "/actions/collection/remove":
		err = a.backend.RemoveCollection(r.Context(), id)
	case "/actions/wishlist/add":
		err = a.backend.AddWishlist(r.Context(), id)
	case "/actions/wishlist/remove":
		err = a.backend.RemoveWishlist(r.Context(), id)
	case "/actions/shelf/add":
		err = a.backend.AddShelf(r.Context(), id)
	case "/actions/shelf/remove":
		err = a.backend.RemoveShelf(r.Context(), id)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		if a.renderNextFigurineCard(w, r, id, next) {
			return
		}
		a.renderNextPageShell(w, r, next)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (a App) renderNextFigurineCard(w http.ResponseWriter, r *http.Request, id, next string) bool {
	u, err := url.ParseRequestURI(next)
	if err != nil || u.Path != "/search" {
		return false
	}

	query := u.Query()
	items, err := a.backend.Figurines(r.Context(), query.Get("q"), query.Get("series_id"), query.Get("ip"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return true
	}
	for _, item := range items {
		if item.ID == id {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = a.templates("search.html").ExecuteTemplate(w, "figurine_card", map[string]any{
				"Figurine": item,
				"Active":   "search",
				"Next":     next,
			})
			return true
		}
	}
	w.WriteHeader(http.StatusNoContent)
	return true
}

func (a App) renderNextPageShell(w http.ResponseWriter, r *http.Request, next string) {
	page, data := a.pageForNext(r, next)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = a.templates(page).ExecuteTemplate(w, "page_shell", data)
}

func (a App) pageForNext(r *http.Request, next string) (string, PageData) {
	u, err := url.ParseRequestURI(next)
	if err != nil || u.Path == "" {
		u = &url.URL{Path: "/"}
	}

	switch u.Path {
	case "/collection":
		items, err := a.backend.Collection(r.Context())
		data := PageData{Title: "My Collection", Active: "collection", Figurines: items, Next: "/collection"}
		if err != nil {
			data.Error = err.Error()
		}
		return "collection.html", data
	case "/search":
		nextReq := r.Clone(r.Context())
		nextReq.URL = u
		return "search.html", a.searchData(nil, nextReq)
	case "/wishlist":
		items, err := a.backend.Wishlist(r.Context())
		data := PageData{Title: "Wishlist", Active: "wishlist", Figurines: items, Next: "/wishlist"}
		if err != nil {
			data.Error = err.Error()
		}
		return "wishlist.html", data
	default:
		items, err := a.backend.Shelf(r.Context())
		data := PageData{Title: "Shelf", Active: "shelf", Figurines: items, Next: "/"}
		if err != nil {
			data.Error = err.Error()
		}
		return "shelf.html", data
	}
}

func (a App) static(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name != "static/styles.css" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	data, err := assets.Files.ReadFile(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(data)
}

func (a App) render(w http.ResponseWriter, page string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = a.templates(page).ExecuteTemplate(w, "layout", data)
}

func (a App) templates(page string) *template.Template {
	t, err := template.New("layout.html").Funcs(template.FuncMap{
		"eq": func(a, b string) bool { return a == b },
		"dict": func(fig backend.Figurine, active, next string) map[string]any {
			return map[string]any{"Figurine": fig, "Active": active, "Next": next}
		},
	}).ParseFS(assets.Files,
		"templates/layout.html",
		"templates/"+page,
		"templates/partials/*.html",
	)
	if err != nil {
		panic(err)
	}
	return t
}

func ips(series []backend.Series) []string {
	seen := map[string]bool{}
	for _, item := range series {
		if item.IP != "" {
			seen[item.IP] = true
		}
	}
	var out []string
	for item := range seen {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
