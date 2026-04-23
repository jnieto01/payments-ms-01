-- +goose Up

ALTER TABLE payments
    ADD COLUMN payment_method VARCHAR(50) NULL COMMENT 'Payment method: transfer, mercado_pago' AFTER item_id;
