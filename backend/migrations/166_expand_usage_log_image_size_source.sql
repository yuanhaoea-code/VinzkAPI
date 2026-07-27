-- Keep the usage log constraint aligned with the image billing resolver.
-- `requested` records an explicit requested tier, while `output_downgrade`
-- records that the returned image was billed below that requested tier.

ALTER TABLE usage_logs
    DROP CONSTRAINT IF EXISTS usage_logs_image_size_source_check;

ALTER TABLE usage_logs
    ADD CONSTRAINT usage_logs_image_size_source_check
    CHECK (
        image_size_source IS NULL
        OR image_size_source IN (
            'output',
            'input',
            'default',
            'legacy',
            'requested',
            'output_downgrade'
        )
    ) NOT VALID;
