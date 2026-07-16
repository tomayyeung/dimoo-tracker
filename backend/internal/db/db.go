package db

import (
	"context"
	"errors"
	"os"
	"sync"

	"dimoo-tracker-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool
	once sync.Once
	err  error
)

// Pool(ctx) creates a shared pgxpool.Pool using DATABASE_URL environment variable.
// Serverless functions should not open a new raw DB connection per query.
// Handles pooling and ensures the pool is initialized once per process/runtime instance.
func Pool(ctx context.Context) (*pgxpool.Pool, error) {
	once.Do(func() {
		url := os.Getenv("DATABASE_URL")
		if url == "" {
			err = errors.New("DATABASE_URL is not set")
			return
		}
		pool, err = pgxpool.New(ctx, url)
	})
	return pool, err
}

// Returns all series ordered by name.
func Series(ctx context.Context) ([]models.Series, error) {
	p, err := Pool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.Query(ctx, `SELECT id::text, name, theme, COALESCE(release_year, 0) FROM series ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Series
	for rows.Next() {
		var item models.Series
		if err := rows.Scan(&item.ID, &item.Name, &item.Theme, &item.ReleaseYear); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Returns catalog figurines.
//
// Supports search by name, series name, or character.
// Supports filtering by series_id.
// Supports filtering by exact character.
//
// Uses EXISTS subqueries to compute owned, wishlisted, and on_shelf.
func Figurines(ctx context.Context, q, seriesID, character string) ([]models.Figurine, error) {
	p, err := Pool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.Query(ctx, `
		SELECT f.id::text, f.series_id::text, s.name, f.name, f.character, f.rarity, f.image_url,
		       EXISTS (SELECT 1 FROM collection_items c WHERE c.figurine_id = f.id) AS owned,
		       EXISTS (SELECT 1 FROM wishlist_items w WHERE w.figurine_id = f.id) AS wishlisted,
		       EXISTS (SELECT 1 FROM shelf_items sh WHERE sh.figurine_id = f.id) AS on_shelf
		FROM figurines f
		JOIN series s ON s.id = f.series_id
		WHERE ($1 = '' OR f.name ILIKE '%' || $1 || '%' OR s.name ILIKE '%' || $1 || '%' OR f.character ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR f.series_id::text = $2)
		  AND ($3 = '' OR f.character = $3)
		ORDER BY s.name, f.name`, q, seriesID, character)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Figurine
	for rows.Next() {
		var item models.Figurine
		if err := rows.Scan(&item.ID, &item.SeriesID, &item.SeriesName, &item.Name, &item.Character, &item.Rarity, &item.ImageURL, &item.Owned, &item.Wishlisted, &item.OnShelf); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Returns all owned figurines.
func Collection(ctx context.Context) ([]models.Figurine, error) {
	return list(ctx, "collection_items", "c.acquired_at")
}

// Returns all wishlisted figurines.
func Wishlist(ctx context.Context) ([]models.Figurine, error) {
	return list(ctx, "wishlist_items", "c.added_at")
}

// Returns featured shelf figurines ordered by position.
func Shelf(ctx context.Context) ([]models.Figurine, error) {
	return list(ctx, "shelf_items", "c.position, c.added_at")
}

// Inserts into collection_items. Uses ON CONFLICT DO NOTHING, so repeated adds are safe.
func AddCollection(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, `INSERT INTO collection_items (figurine_id) VALUES ($1) ON CONFLICT DO NOTHING`, id)
	return err
}

// Deletes from shelf_items first, then collection_items.
// Wrapped in a transaction so a figurine cannot remain on the shelf after ownership is removed.
func RemoveCollection(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM shelf_items WHERE figurine_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM collection_items WHERE figurine_id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Inserts into wishlist_items.
func AddWishlist(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, `INSERT INTO wishlist_items (figurine_id) VALUES ($1) ON CONFLICT DO NOTHING`, id)
	return err
}

// Deletes from wishlist_items.
func RemoveWishlist(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, `DELETE FROM wishlist_items WHERE figurine_id = $1`, id)
	return err
}

// Adds the figurine to collection_items first, then inserts into shelf_items.
//
// Assigns position using MAX(position) + 1.
//
// Wrapped in a transaction so shelf and collection state stay consistent.
func AddShelf(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO collection_items (figurine_id) VALUES ($1) ON CONFLICT DO NOTHING`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO shelf_items (figurine_id, position)
		VALUES ($1, COALESCE((SELECT MAX(position) + 1 FROM shelf_items), 1))
		ON CONFLICT DO NOTHING`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Deletes only from shelf_items.
func RemoveShelf(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, `DELETE FROM shelf_items WHERE figurine_id = $1`, id)
	return err
}

func Characters(ctx context.Context) ([]string, error) {
	p, err := Pool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.Query(ctx, `SELECT DISTINCT character FROM figurines ORDER BY character`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Joins the selected item table to figurines and series, then returns the same enriched Figurine shape.
func list(ctx context.Context, table, order string) ([]models.Figurine, error) {
	p, err := Pool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.Query(ctx, `
		SELECT f.id::text, f.series_id::text, s.name, f.name, f.character, f.rarity, f.image_url,
		       EXISTS (SELECT 1 FROM collection_items ci WHERE ci.figurine_id = f.id) AS owned,
		       EXISTS (SELECT 1 FROM wishlist_items wi WHERE wi.figurine_id = f.id) AS wishlisted,
		       EXISTS (SELECT 1 FROM shelf_items si WHERE si.figurine_id = f.id) AS on_shelf
		FROM `+table+` c
		JOIN figurines f ON f.id = c.figurine_id
		JOIN series s ON s.id = f.series_id
		ORDER BY `+order+`, s.name, f.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Figurine
	for rows.Next() {
		var item models.Figurine
		if err := rows.Scan(&item.ID, &item.SeriesID, &item.SeriesName, &item.Name, &item.Character, &item.Rarity, &item.ImageURL, &item.Owned, &item.Wishlisted, &item.OnShelf); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
