package models

// Represents a Pop Mart series/set.
type Series struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IP          string `json:"ip"` // Intellectual property (Dimoo, Minions, etc)
	ReleaseYear int    `json:"release_year,omitempty"`
}

// Represents one catalog figurine plus the user-specific state for that figurine.
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

// Used by POST endpoints when adding a figurine to collection, wishlist, or shelf.
type FigurineInput struct {
	FigurineID string `json:"figurine_id"`
}
