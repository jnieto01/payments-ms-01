-- +goose Down

ALTER TABLE payments
    DROP COLUMN payment_method;
