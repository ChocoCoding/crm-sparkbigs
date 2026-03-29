-- ═══════════════════════════════════════════════════════════════
-- Migración 001: Tablas de autenticación y usuarios
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS users (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at      DATETIME(3) NULL,
    updated_at      DATETIME(3) NULL,
    deleted_at      DATETIME(3) NULL,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    role            VARCHAR(20)  NOT NULL DEFAULT 'user',
    is_active       TINYINT(1)   NOT NULL DEFAULT 1,
    must_change_pass TINYINT(1)  NOT NULL DEFAULT 1,
    last_login_at   DATETIME(3)  NULL,
    encrypted_api_key VARCHAR(512) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY idx_users_email (email),
    KEY idx_users_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS licenses (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    plan       VARCHAR(50)     NOT NULL DEFAULT 'free',
    is_active  TINYINT(1)      NOT NULL DEFAULT 1,
    expires_at DATETIME(3)     NULL,
    PRIMARY KEY (id),
    UNIQUE KEY idx_licenses_user_id (user_id),
    KEY idx_licenses_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME(3) NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    token      VARCHAR(512)    NOT NULL,
    expires_at DATETIME(3)     NOT NULL,
    revoked    TINYINT(1)      NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY idx_refresh_tokens_token (token),
    KEY idx_refresh_tokens_user_id (user_id)
);
