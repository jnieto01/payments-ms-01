-- +goose Up

ALTER TABLE plans
    ADD COLUMN short_name VARCHAR(20) NOT NULL DEFAULT '' AFTER name;

UPDATE plans SET short_name = 'PRO' WHERE plan_key = 'pro';
UPDATE plans SET short_name = 'AVZ' WHERE plan_key = 'avanzado';
