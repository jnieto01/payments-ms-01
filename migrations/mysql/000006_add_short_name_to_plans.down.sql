-- +goose Down

ALTER TABLE plans DROP COLUMN short_name;
