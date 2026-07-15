-- Records manual card-code purchases fulfilled through Liandong Xiaopu.
-- The first version has no payment callback, so rows are created when a
-- matching redeem code is successfully redeemed.

CREATE TABLE IF NOT EXISTS ldxp_recharge_orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    user_email VARCHAR(255) NOT NULL DEFAULT '',
    source VARCHAR(50) NOT NULL DEFAULT 'ldxp',
    product_name VARCHAR(255) NOT NULL DEFAULT '',
    product_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    balance_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    redeem_code_masked VARCHAR(64) NOT NULL DEFAULT '',
    ldxp_order_no VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(30) NOT NULL DEFAULT 'redeemed',
    paid_at TIMESTAMPTZ,
    redeemed_at TIMESTAMPTZ,
    client_ip VARCHAR(50) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ldxp_recharge_orders_redeem_code_id_unique
    ON ldxp_recharge_orders(redeem_code_id)
    WHERE redeem_code_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ldxp_recharge_orders_user_id
    ON ldxp_recharge_orders(user_id);

CREATE INDEX IF NOT EXISTS idx_ldxp_recharge_orders_created_at
    ON ldxp_recharge_orders(created_at);

CREATE INDEX IF NOT EXISTS idx_ldxp_recharge_orders_redeemed_at
    ON ldxp_recharge_orders(redeemed_at);

CREATE INDEX IF NOT EXISTS idx_ldxp_recharge_orders_status
    ON ldxp_recharge_orders(status);
