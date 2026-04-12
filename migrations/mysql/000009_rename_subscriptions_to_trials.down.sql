-- +goose Down

RENAME TABLE trials TO subscriptions;
