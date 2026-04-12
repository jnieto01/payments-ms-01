-- +goose Up
ALTER TABLE plans
    ADD COLUMN mp_preapproval_plan_id  VARCHAR(255) NULL COMMENT 'MercadoPago /preapproval_plan ID',
    ADD COLUMN mp_preapproval_plan_url TEXT         NULL COMMENT 'MercadoPago init_point to subscribe';

ALTER TABLE subscriptions
    ADD COLUMN mp_preapproval_id     VARCHAR(255) NULL COMMENT 'MercadoPago /preapproval ID (per subscriber)',
    ADD COLUMN mp_preapproval_status VARCHAR(50)  NULL COMMENT 'authorized | paused | cancelled';
