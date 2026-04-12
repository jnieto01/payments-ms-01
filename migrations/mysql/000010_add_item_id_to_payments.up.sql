-- +goose Up

ALTER TABLE payments
    ADD COLUMN item_id BIGINT NULL COMMENT 'Market item ID for advertising payments',
    ADD INDEX  idx_payments_item_id (item_id);
