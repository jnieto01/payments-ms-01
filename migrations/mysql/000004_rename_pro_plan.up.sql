-- +goose Up

UPDATE plans SET name = 'Profesional' WHERE plan_key = 'pro';
