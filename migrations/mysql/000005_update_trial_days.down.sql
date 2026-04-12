-- +goose Down

UPDATE plans SET trial_days = 14 WHERE plan_key = 'avanzado';
UPDATE plans SET trial_days = 14 WHERE plan_key = 'pro';
