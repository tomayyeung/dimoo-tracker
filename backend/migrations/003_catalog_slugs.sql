ALTER TABLE series ADD COLUMN IF NOT EXISTS slug TEXT;
ALTER TABLE series ADD COLUMN IF NOT EXISTS image_path TEXT NOT NULL DEFAULT '';

ALTER TABLE figurines ADD COLUMN IF NOT EXISTS slug TEXT;
ALTER TABLE figurines ADD COLUMN IF NOT EXISTS image_path TEXT NOT NULL DEFAULT '';

UPDATE series
SET slug = lower(regexp_replace(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g'), '(^-|-$)', '', 'g'))
WHERE slug IS NULL OR slug = '';

UPDATE figurines
SET slug = lower(regexp_replace(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g'), '(^-|-$)', '', 'g'))
WHERE slug IS NULL OR slug = '';

UPDATE series SET image_path = slug WHERE image_path = '';
UPDATE figurines SET image_path = slug || '.webp' WHERE image_path = '';

ALTER TABLE series ALTER COLUMN slug SET NOT NULL;
ALTER TABLE figurines ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_series_slug ON series(slug);
CREATE UNIQUE INDEX IF NOT EXISTS idx_figurines_series_slug ON figurines(series_id, slug);
