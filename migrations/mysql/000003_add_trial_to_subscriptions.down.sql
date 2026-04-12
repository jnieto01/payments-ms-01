-- +goose Down

ALTER TABLE subscriptions
    DROP COLUMN trial_ends_at,
    DROP COLUMN trial_starts_at,
    DROP COLUMN is_trial;
