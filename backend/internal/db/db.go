// Package db provides PostgreSQL-backed catalog and single-user collection
// state queries for the backend API.
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
	pool        *pgxpool.Pool
	once        sync.Once
	err         error
	ErrNotOwned = errors.New("figurine is not owned")
)

// Pool creates a shared pgxpool.Pool using the DATABASE_URL environment variable.
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

// Series returns all catalog series ordered by name.
func Series(ctx context.Context) ([]models.Series, error) {
	p, err := Pool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.Query(ctx, `SELECT id::text, name, ip, COALESCE(release_year, 0) FROM series ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Series
	for rows.Next() {
		var item models.Series
		if err := rows.Scan(&item.ID, &item.Name, &item.IP, &item.ReleaseYear); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Figurines returns catalog figurines matching the optional filters.
//
// Supports search by name, series name, or intellectual property.
// Supports filtering by series_id.
// Supports filtering by exact intellectual property.
//
// Uses EXISTS subqueries to compute owned, wishlisted, and on_shelf.
func Figurines(ctx context.Context, q, seriesID, ip string) ([]models.Figurine, error) {
	p, err := Pool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.Query(ctx, `
		SELECT f.id::text, f.series_id::text, s.name, f.name, f.rarity, f.image_url,
		       EXISTS (SELECT 1 FROM collection_items c WHERE c.figurine_id = f.id) AS owned,
		       EXISTS (SELECT 1 FROM wishlist_items w WHERE w.figurine_id = f.id) AS wishlisted,
		       EXISTS (SELECT 1 FROM shelf_items sh WHERE sh.figurine_id = f.id) AS on_shelf
		FROM figurines f
		JOIN series s ON s.id = f.series_id
		WHERE ($1 = '' OR f.name ILIKE '%' || $1 || '%' OR s.name ILIKE '%' || $1 || '%' OR s.ip ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR f.series_id::text = $2)
		  AND ($3 = '' OR s.ip = $3)
		ORDER BY s.name, f.name`, q, seriesID, ip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Figurine
	for rows.Next() {
		var item models.Figurine
		if err := rows.Scan(&item.ID, &item.SeriesID, &item.SeriesName, &item.Name, &item.Rarity, &item.ImageURL, &item.Owned, &item.Wishlisted, &item.OnShelf); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Collection returns all owned figurines.
func Collection(ctx context.Context) ([]models.Figurine, error) {
	return list(ctx, "collection_items", "c.acquired_at")
}

// Wishlist returns all wishlisted figurines.
func Wishlist(ctx context.Context) ([]models.Figurine, error) {
	return list(ctx, "wishlist_items", "c.added_at")
}

// Shelf returns featured shelf figurines ordered by position.
func Shelf(ctx context.Context) ([]models.Figurine, error) {
	return list(ctx, "shelf_items", "c.position, c.added_at")
}

// AddCollection inserts a figurine into collection_items and removes it from wishlist_items.
//
// It uses ON CONFLICT DO NOTHING, so repeated adds are safe.
func AddCollection(ctx context.Context, id string) error {
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
	if _, err := tx.Exec(ctx, `DELETE FROM wishlist_items WHERE figurine_id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RemoveCollection deletes a figurine from shelf_items first, then collection_items.
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

// AddWishlist inserts a figurine into wishlist_items.
func AddWishlist(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, `INSERT INTO wishlist_items (figurine_id) VALUES ($1) ON CONFLICT DO NOTHING`, id)
	return err
}

// RemoveWishlist deletes a figurine from wishlist_items.
func RemoveWishlist(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, `DELETE FROM wishlist_items WHERE figurine_id = $1`, id)
	return err
}

// AddShelf adds an owned figurine to shelf_items and removes it from wishlist_items.
//
// Assigns position using MAX(position) + 1.
//
// Wrapped in a transaction so shelf and wishlist state stay consistent.
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
	cmd, err := tx.Exec(ctx, `
		INSERT INTO shelf_items (figurine_id, position)
		SELECT $1, COALESCE((SELECT MAX(position) + 1 FROM shelf_items), 1)
		WHERE EXISTS (SELECT 1 FROM collection_items WHERE figurine_id = $1)
		ON CONFLICT DO NOTHING`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		var owned bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM collection_items WHERE figurine_id = $1)`, id).Scan(&owned); err != nil {
			return err
		}
		if !owned {
			return ErrNotOwned
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM wishlist_items WHERE figurine_id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RemoveShelf deletes a figurine only from shelf_items.
func RemoveShelf(ctx context.Context, id string) error {
	p, err := Pool(ctx)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, `DELETE FROM shelf_items WHERE figurine_id = $1`, id)
	return err
}

// IPs returns distinct non-empty intellectual property names from catalog series.
func IPs(ctx context.Context) ([]string, error) {
	p, err := Pool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.Query(ctx, `SELECT DISTINCT ip FROM series WHERE ip <> '' ORDER BY ip`)
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
		SELECT f.id::text, f.series_id::text, s.name, f.name, f.rarity, f.image_url,
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
		if err := rows.Scan(&item.ID, &item.SeriesID, &item.SeriesName, &item.Name, &item.Rarity, &item.ImageURL, &item.Owned, &item.Wishlisted, &item.OnShelf); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
