-- +goose Up

-- Avanzado: 30 días de prueba (incluye funcionalidades de ambos planes)
UPDATE plans SET trial_days = 30 WHERE plan_key = 'avanzado';

-- Profesional: sin período de prueba
UPDATE plans SET trial_days = 0 WHERE plan_key = 'pro';
