INSERT INTO series (id, name, theme, release_year) VALUES
  ('00000000-0000-0000-0000-000000000101', 'Dimoo Dream Travel', 'soft clouds, luggage, sleepy adventures', 2023),
  ('00000000-0000-0000-0000-000000000102', 'Dimoo Animal Kingdom', 'cozy animal costumes and woodland friends', 2022),
  ('00000000-0000-0000-0000-000000000103', 'Dimoo Space Journey', 'stars, planets, and astronaut daydreams', 2024),
  ('00000000-0000-0000-0000-000000000104', 'Dimoo Cafe Time', 'desserts, drinks, and relaxed afternoons', 2024)
ON CONFLICT (name) DO NOTHING;

INSERT INTO figurines (id, series_id, name, character, rarity) VALUES
  ('00000000-0000-0000-0000-000000001001', '00000000-0000-0000-0000-000000000101', 'Cloud Boarding Pass', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001002', '00000000-0000-0000-0000-000000000101', 'Moonlit Suitcase', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001003', '00000000-0000-0000-0000-000000000101', 'Secret Window Seat', 'Dimoo', 'secret'),
  ('00000000-0000-0000-0000-000000001004', '00000000-0000-0000-0000-000000000102', 'Bear Hoodie Nap', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001005', '00000000-0000-0000-0000-000000000102', 'Fox Lantern Walk', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001006', '00000000-0000-0000-0000-000000000102', 'Bunny Meadow', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001007', '00000000-0000-0000-0000-000000000103', 'Tiny Astronaut', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001008', '00000000-0000-0000-0000-000000000103', 'Planet Collector', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001009', '00000000-0000-0000-0000-000000000103', 'Galaxy Secret', 'Dimoo', 'secret'),
  ('00000000-0000-0000-0000-000000001010', '00000000-0000-0000-0000-000000000104', 'Latte Foam Friend', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001011', '00000000-0000-0000-0000-000000000104', 'Strawberry Cake Seat', 'Dimoo', 'standard'),
  ('00000000-0000-0000-0000-000000001012', '00000000-0000-0000-0000-000000000104', 'Midnight Macaron', 'Dimoo', 'standard')
ON CONFLICT (series_id, name) DO NOTHING;
