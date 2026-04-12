-- +goose Up

CREATE TABLE IF NOT EXISTS marketplace_commissions (
    id          INT             AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100)    NOT NULL                    COMMENT 'Internal key, e.g. marketplace_listing',
    percentage  DECIMAL(5,4)    NOT NULL DEFAULT 0.0200     COMMENT 'Commission rate, e.g. 0.0200 = 2%',
    wording_es  TEXT            NOT NULL                    COMMENT 'Display text shown to user in Spanish',
    wording_en  TEXT            NOT NULL                    COMMENT 'Display text shown to user in English',
    currency    VARCHAR(10)     NOT NULL DEFAULT 'ARS',
    is_active   TINYINT(1)      NOT NULL DEFAULT 1,
    created_at  DATETIME        NOT NULL DEFAULT NOW(),
    updated_at  DATETIME        NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    UNIQUE KEY idx_marketplace_commissions_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO marketplace_commissions (name, percentage, wording_es, wording_en, currency)
VALUES (
    'marketplace_listing',
    0.0200,
    'Al publicar se cobra una comisión del 2% sobre el precio de venta. El pago se realiza a través de MercadoPago.',
    'A 2% commission on the sale price is charged upon listing. Payment is processed through MercadoPago.',
    'ARS'
);
