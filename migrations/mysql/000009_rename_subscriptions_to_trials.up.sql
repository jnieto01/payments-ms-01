-- +goose Up

RENAME TABLE subscriptions TO trials;
