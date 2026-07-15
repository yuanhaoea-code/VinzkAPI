ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS image_allowed_tiers JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE groups
SET image_allowed_tiers = '["1K"]'::jsonb
WHERE allow_image_generation = true
  AND image_allowed_tiers = '[]'::jsonb;

ALTER TABLE groups
    DROP CONSTRAINT IF EXISTS groups_image_allowed_tiers_valid;

ALTER TABLE groups
    ADD CONSTRAINT groups_image_allowed_tiers_valid
    CHECK (
        jsonb_typeof(image_allowed_tiers) = 'array'
        AND image_allowed_tiers <@ '["1K", "2K", "4K"]'::jsonb
    );

COMMENT ON COLUMN groups.image_allowed_tiers IS 'Allowed image product tiers: 1K, 2K, 4K';

ALTER TABLE image_generations
    ADD COLUMN IF NOT EXISTS resolution_tier VARCHAR(8) NOT NULL DEFAULT '1K',
    ADD COLUMN IF NOT EXISTS aspect_ratio VARCHAR(16) NOT NULL DEFAULT '1:1';

UPDATE image_generations
SET resolution_tier = CASE
        WHEN UPPER(TRIM(size)) = '4K' OR LOWER(TRIM(size)) IN ('2880x2880', '3840x2160', '2160x3840') THEN '4K'
        WHEN UPPER(TRIM(size)) = '2K' OR LOWER(TRIM(size)) IN ('2048x2048', '2048x1152', '1152x2048') THEN '2K'
        ELSE '1K'
    END,
    aspect_ratio = CASE
        WHEN LOWER(TRIM(size)) IN ('1536x1024', '2048x1365', '2880x1920', '3840x2160') THEN '3:2'
        WHEN LOWER(TRIM(size)) IN ('1024x1536', '1365x2048', '1920x2880', '2160x3840') THEN '2:3'
        ELSE '1:1'
    END
WHERE resolution_tier = '1K'
  AND aspect_ratio = '1:1';
