// Package models defines the JSON shapes shared by backend API handlers and
// database access code.
package models

// Series represents a Pop Mart series or set in the catalog.
type Series struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IP          string `json:"ip"` // Intellectual property (Dimoo, Minions, etc)
	ReleaseYear int    `json:"release_year,omitempty"`
}

// Figurine represents one catalog figurine plus the user's state for it.
type Figurine struct {
	ID         string `json:"id"`
	SeriesID   string `json:"series_id"`
	SeriesName string `json:"series_name"`
	Name       string `json:"name"`
	Rarity     string `json:"rarity"`
	ImageURL   string `json:"image_url"`
	Owned      bool   `json:"owned"`
	Wishlisted bool   `json:"wishlisted"`
	OnShelf    bool   `json:"on_shelf"`
}

// FigurineInput is the POST body for collection, wishlist, and shelf mutations.
type FigurineInput struct {
	FigurineID string `json:"figurine_id"`
}
