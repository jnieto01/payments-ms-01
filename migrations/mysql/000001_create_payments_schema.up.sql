-- +goose Up

CREATE TABLE IF NOT EXISTS payments (
    id              INT             AUTO_INCREMENT PRIMARY KEY,
    club_id         VARCHAR(100)    NOT NULL,
    user_id         VARCHAR(100)    NOT NULL,
    type            VARCHAR(50)     NOT NULL,
    plan            VARCHAR(50),
    amount          DECIMAL(12,2)   NOT NULL,
    currency        VARCHAR(10)     NOT NULL DEFAULT 'ARS',
    status          VARCHAR(50)     NOT NULL DEFAULT 'pending',
    idempotency_key VARCHAR(100)    NOT NULL,
    mp_payment_id   BIGINT,
    mp_preference_id VARCHAR(255),
    init_point      TEXT,
    external_ref    VARCHAR(255),
    created_at      DATETIME        NOT NULL DEFAULT NOW(),
    updated_at      DATETIME        NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    UNIQUE KEY idx_payments_idempotency_key (idempotency_key),
    INDEX idx_payments_club_id (club_id),
    INDEX idx_payments_mp_payment_id (mp_payment_id),
    INDEX idx_payments_external_ref (external_ref)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS subscriptions (
    id          INT             AUTO_INCREMENT PRIMARY KEY,
    club_id     VARCHAR(100)    NOT NULL,
    plan        VARCHAR(50)     NOT NULL,
    status      VARCHAR(50)     NOT NULL DEFAULT 'active',
    payment_id  INT,
    starts_at   DATETIME        NOT NULL,
    ends_at     DATETIME        NOT NULL,
    created_at  DATETIME        NOT NULL DEFAULT NOW(),
    updated_at  DATETIME        NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    UNIQUE KEY idx_subscriptions_club_id (club_id),
    CONSTRAINT fk_sub_payment FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
