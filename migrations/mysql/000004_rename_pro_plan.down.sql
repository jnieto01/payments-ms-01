-- +goose Down

UPDATE plans SET name = 'Pro' WHERE plan_key = 'pro';
