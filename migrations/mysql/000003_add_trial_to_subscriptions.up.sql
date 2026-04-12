-- +goose Up

ALTER TABLE subscriptions
    ADD COLUMN is_trial        TINYINT(1)  NOT NULL DEFAULT 0  AFTER status,
    ADD COLUMN trial_starts_at DATETIME                        AFTER is_trial,
    ADD COLUMN trial_ends_at   DATETIME                        AFTER trial_starts_at;
