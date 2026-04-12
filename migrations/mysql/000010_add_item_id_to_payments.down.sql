-- +goose Down

ALTER TABLE payments
    DROP INDEX idx_payments_item_id,
    DROP COLUMN item_id;
