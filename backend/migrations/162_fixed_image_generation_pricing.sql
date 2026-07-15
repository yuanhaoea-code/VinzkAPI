-- Image studio uses two exclusive key pools and a fixed per-image price list.
UPDATE groups
SET image_allowed_tiers = CASE
        WHEN image_allowed_tiers ? '2K' AND image_allowed_tiers ? '4K' THEN '["2K", "4K"]'::jsonb
        WHEN image_allowed_tiers ? '4K' THEN '["4K"]'::jsonb
        WHEN image_allowed_tiers ? '2K' THEN '["2K"]'::jsonb
        ELSE '["1K"]'::jsonb
    END,
    image_rate_independent = true,
    image_rate_multiplier = 1,
    image_price_1k = 0.06,
    image_price_2k = 0.16,
    image_price_4k = 0.20
WHERE allow_image_generation = true;

ALTER TABLE groups
    DROP CONSTRAINT IF EXISTS groups_image_allowed_tiers_pool_exclusive;

ALTER TABLE groups
    ADD CONSTRAINT groups_image_allowed_tiers_pool_exclusive
    CHECK (
        allow_image_generation = false
        OR NOT (
            image_allowed_tiers ? '1K'
            AND (image_allowed_tiers ? '2K' OR image_allowed_tiers ? '4K')
        )
    );

COMMENT ON CONSTRAINT groups_image_allowed_tiers_pool_exclusive ON groups IS
    'Image key groups are either standard (1K) or HD (2K/4K), never both.';
