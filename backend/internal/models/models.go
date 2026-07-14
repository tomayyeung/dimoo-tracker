package models

type Series struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Theme       string `json:"theme"`
	ReleaseYear int    `json:"release_year,omitempty"`
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

type FigurineInput struct {
	FigurineID string `json:"figurine_id"`
}
