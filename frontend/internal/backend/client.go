package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

type Series struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Theme       string `json:"theme"`
	ReleaseYear int    `json:"release_year"`
}

type Figurine struct {
	ID         string `json:"id"`
	SeriesID   string `json:"series_id"`
	SeriesName string `json:"series_name"`
	Name       string `json:"name"`
	Character  string `json:"character"`
	Rarity     string `json:"rarity"`
	ImageURL   string `json:"image_url"`
	Owned      bool   `json:"owned"`
	Wishlisted bool   `json:"wishlisted"`
	OnShelf    bool   `json:"on_shelf"`
}

func New() Client {
	base := strings.TrimRight(os.Getenv("API_BASE_URL"), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return Client{BaseURL: base, HTTP: &http.Client{Timeout: 8 * time.Second}}
}

func (c Client) Series(ctx context.Context) ([]Series, error) {
	var items []Series
	return items, c.get(ctx, "/api/series", &items)
}

func (c Client) Figurines(ctx context.Context, q, seriesID, character string) ([]Figurine, error) {
	values := url.Values{}
	values.Set("q", q)
	values.Set("series_id", seriesID)
	values.Set("character", character)
	var items []Figurine
	return items, c.get(ctx, "/api/figurines?"+values.Encode(), &items)
}

func (c Client) Collection(ctx context.Context) ([]Figurine, error) {
	var items []Figurine
	return items, c.get(ctx, "/api/collection", &items)
}

func (c Client) Wishlist(ctx context.Context) ([]Figurine, error) {
	var items []Figurine
	return items, c.get(ctx, "/api/wishlist", &items)
}

func (c Client) Shelf(ctx context.Context) ([]Figurine, error) {
	var items []Figurine
	return items, c.get(ctx, "/api/shelf", &items)
}

func (c Client) AddCollection(ctx context.Context, id string) error {
	return c.postID(ctx, "/api/collection", id)
}

func (c Client) RemoveCollection(ctx context.Context, id string) error {
	return c.deleteID(ctx, "/api/collection", id)
}

func (c Client) AddWishlist(ctx context.Context, id string) error {
	return c.postID(ctx, "/api/wishlist", id)
}

func (c Client) RemoveWishlist(ctx context.Context, id string) error {
	return c.deleteID(ctx, "/api/wishlist", id)
}

func (c Client) AddShelf(ctx context.Context, id string) error {
	return c.postID(ctx, "/api/shelf", id)
}

func (c Client) RemoveShelf(ctx context.Context, id string) error {
	return c.deleteID(ctx, "/api/shelf", id)
}

func (c Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c Client) postID(ctx context.Context, path, id string) error {
	body, _ := json.Marshal(map[string]string{"figurine_id": id})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, nil)
}

func (c Client) deleteID(ctx context.Context, path, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+path+"?id="+url.QueryEscape(id), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

func (c Client) do(req *http.Request, out any) error {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var payload struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &payload); err == nil && payload.Error != "" {
			return fmt.Errorf("backend returned %s: %s", resp.Status, payload.Error)
		}
		if len(body) > 0 {
			return fmt.Errorf("backend returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}
		return fmt.Errorf("backend returned %s", resp.Status)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
