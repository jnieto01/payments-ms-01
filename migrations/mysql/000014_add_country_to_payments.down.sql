-- +goose Down

ALTER TABLE payments
    DROP INDEX idx_payments_country,
    DROP COLUMN country;
