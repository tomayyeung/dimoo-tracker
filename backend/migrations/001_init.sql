CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS series (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  theme TEXT NOT NULL DEFAULT '',
  release_year INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS figurines (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  character TEXT NOT NULL,
  rarity TEXT NOT NULL DEFAULT 'standard',
  image_url TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (series_id, name)
);

CREATE TABLE IF NOT EXISTS collection_items (
  figurine_id UUID PRIMARY KEY REFERENCES figurines(id) ON DELETE CASCADE,
  acquired_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wishlist_items (
  figurine_id UUID PRIMARY KEY REFERENCES figurines(id) ON DELETE CASCADE,
  added_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shelf_items (
  figurine_id UUID PRIMARY KEY REFERENCES figurines(id) ON DELETE CASCADE,
  position INTEGER NOT NULL DEFAULT 0,
  added_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_figurines_series_id ON figurines(series_id);
CREATE INDEX IF NOT EXISTS idx_figurines_character ON figurines(character);
