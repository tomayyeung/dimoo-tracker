package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Catalog struct {
	StorageBaseURL string          `json:"storage_base_url"`
	Series         []catalogSeries `json:"series"`
}

type catalogSeries struct {
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	IP          string            `json:"ip"`
	ReleaseYear int               `json:"release_year"`
	ImagePath   string            `json:"image_path"`
	Figurines   []catalogFigurine `json:"figurines"`
}

type catalogFigurine struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Rarity    string `json:"rarity"`
	ImagePath string `json:"image_path"`
}

func main() {
	loadDotEnv()

	file := "./catalog.json"
	if len(os.Args) > 1 {
		file = os.Args[1]
	}

	catalog, err := readCatalog(file)
	if err != nil {
		log.Fatal(err)
	}
	if err := validateCatalog(catalog); err != nil {
		log.Fatal(err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := importCatalog(ctx, pool, catalog); err != nil {
		log.Fatal(err)
	}
	log.Printf("imported %d series", len(catalog.Series))
}

func readCatalog(file string) (Catalog, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Catalog{}, err
	}
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

func validateCatalog(catalog Catalog) error {
	if strings.TrimSpace(catalog.StorageBaseURL) == "" {
		return errors.New("storage_base_url is required")
	}
	if len(catalog.Series) == 0 {
		return errors.New("at least one series is required")
	}
	seriesSlugs := map[string]bool{}
	for _, series := range catalog.Series {
		if strings.TrimSpace(series.Name) == "" || strings.TrimSpace(series.Slug) == "" {
			return fmt.Errorf("series name and slug are required: %q", series.Name)
		}
		if strings.TrimSpace(series.ImagePath) == "" {
			return fmt.Errorf("series image_path is required: %s", series.Slug)
		}
		if strings.TrimSpace(series.IP) == "" {
			return fmt.Errorf("series ip is required: %s", series.Slug)
		}
		if seriesSlugs[series.Slug] {
			return fmt.Errorf("duplicate series slug: %s", series.Slug)
		}
		seriesSlugs[series.Slug] = true

		figurineSlugs := map[string]bool{}
		for _, figurine := range series.Figurines {
			if strings.TrimSpace(figurine.Name) == "" || strings.TrimSpace(figurine.Slug) == "" {
				return fmt.Errorf("figurine name and slug are required in series %s", series.Slug)
			}
			if strings.TrimSpace(figurine.Rarity) == "" {
				return fmt.Errorf("figurine rarity is required: %s/%s", series.Slug, figurine.Slug)
			}
			if strings.TrimSpace(figurine.ImagePath) == "" {
				return fmt.Errorf("figurine image_path is required: %s/%s", series.Slug, figurine.Slug)
			}
			if figurineSlugs[figurine.Slug] {
				return fmt.Errorf("duplicate figurine slug in %s: %s", series.Slug, figurine.Slug)
			}
			figurineSlugs[figurine.Slug] = true
		}
	}
	return nil
}

func importCatalog(ctx context.Context, pool *pgxpool.Pool, catalog Catalog) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, series := range catalog.Series {
		var seriesID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO series (name, slug, ip, release_year, image_path)
			VALUES ($1, $2, $3, NULLIF($4, 0), $5)
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name,
				ip = EXCLUDED.ip,
				release_year = EXCLUDED.release_year,
				image_path = EXCLUDED.image_path
			RETURNING id::text`, series.Name, series.Slug, series.IP, series.ReleaseYear, series.ImagePath).Scan(&seriesID); err != nil {
			return err
		}

		for _, figurine := range series.Figurines {
			imageURL := buildImageURL(catalog.StorageBaseURL, series.ImagePath, figurine.ImagePath)
			if _, err := tx.Exec(ctx, `
				INSERT INTO figurines (series_id, name, slug, rarity, image_path, image_url)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (series_id, slug) DO UPDATE SET
					name = EXCLUDED.name,
					rarity = EXCLUDED.rarity,
					image_path = EXCLUDED.image_path,
					image_url = EXCLUDED.image_url`, seriesID, figurine.Name, figurine.Slug, figurine.Rarity, figurine.ImagePath, imageURL); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func buildImageURL(baseURL, seriesPath, figurinePath string) string {
	return strings.TrimRight(baseURL, "/") + "/" + trimSlashes(seriesPath) + "/" + trimSlashes(figurinePath)
}

func trimSlashes(value string) string {
	return strings.Trim(strings.TrimSpace(value), "/")
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
