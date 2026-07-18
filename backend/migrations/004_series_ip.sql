ALTER TABLE series ADD COLUMN IF NOT EXISTS ip TEXT NOT NULL DEFAULT '';

UPDATE series s
SET ip = f.ip
FROM (
  SELECT series_id, MIN(character) AS ip
  FROM figurines
  WHERE character <> ''
  GROUP BY series_id
) f
WHERE s.id = f.series_id
  AND s.ip = '';

DROP INDEX IF EXISTS idx_figurines_character;

ALTER TABLE figurines DROP COLUMN IF EXISTS character;
ALTER TABLE series DROP COLUMN IF EXISTS theme;
