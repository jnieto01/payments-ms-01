-- +goose Down
ALTER TABLE subscriptions
    DROP COLUMN mp_preapproval_id,
    DROP COLUMN mp_preapproval_status;

ALTER TABLE plans
    DROP COLUMN mp_preapproval_plan_id,
    DROP COLUMN mp_preapproval_plan_url;
