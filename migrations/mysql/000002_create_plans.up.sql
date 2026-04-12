-- +goose Up

CREATE TABLE IF NOT EXISTS plans (
    id           INT            AUTO_INCREMENT PRIMARY KEY,
    plan_key     VARCHAR(50)    NOT NULL,
    name         VARCHAR(100)   NOT NULL,
    price        DECIMAL(12,2)  NOT NULL,
    currency     VARCHAR(10)    NOT NULL DEFAULT 'ARS',
    period       VARCHAR(50)    NOT NULL DEFAULT 'monthly',
    features     JSON           NOT NULL,
    trial_days   INT            NOT NULL DEFAULT 14,
    warning_days INT            NOT NULL DEFAULT 5,
    is_active    TINYINT(1)     NOT NULL DEFAULT 1,
    created_at   DATETIME       NOT NULL DEFAULT NOW(),
    updated_at   DATETIME       NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    UNIQUE KEY idx_plans_key (plan_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO plans (plan_key, name, price, currency, period, features, trial_days, warning_days) VALUES
(
    'pro',
    'Profesional',
    12000.00,
    'ARS',
    'monthly',
    '["Amistosos entre equipos","Torneos con bracket automático","Gestión de registraciones","Notificaciones a participantes"]',
    0,
    5
),
(
    'avanzado',
    'Avanzado',
    15000.00,
    'ARS',
    'monthly',
    '["Todo el plan Pro","Publicidad en transmisiones","Analytics detallados","Streaming de eventos con ads","Asignación de publicidad por evento","Soporte prioritario"]',
    30,
    5
);
