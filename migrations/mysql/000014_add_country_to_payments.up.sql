-- +goose Up

ALTER TABLE payments
    ADD COLUMN country VARCHAR(2) NOT NULL DEFAULT 'AR' AFTER currency,
    ADD INDEX idx_payments_country (country);
