-- Migration 002: User settings feature (privacy, sessions, deactivate, email change)
--
-- Cac lenh tao bang/cot moi cho tinh nang User Settings:
--   - user_settings                    (1-1 voi users, quyen rieng tu)
--   - user_sessions                    (phiên dang nhap, id = JWT jti)
--   - users.self_deactivated_at        (danh dau tai khoan tu vo hieu hoa)
--   - email_verification_tokens.purpose / pending_email (ho tro doi email)
--
-- Cach chay:
--   mysql -u <user> -p <db_name> < docs/migrations/002_user_settings.sql
-- Hoac paste vao MySQL client / phpMyAdmin.
--
-- LUU Y: MySQL khong ho tro "ADD COLUMN IF NOT EXISTS", nen chi chay mot lan.
-- Khi seed lai toan bo (cmd/seed) thi schema tu dong tao ca cac bang/cot nay.

-- 1. Bang cai dat quyen rieng tu cua user (1-1 voi users)
CREATE TABLE IF NOT EXISTS user_settings (
    user_id VARCHAR(36) PRIMARY KEY,
    discoverable_in_search TINYINT(1) NOT NULL DEFAULT 1,
    allow_stranger_messages TINYINT(1) NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Bang phiên dang nhap (id chinh la JWT jti)
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    device_name VARCHAR(255) NOT NULL DEFAULT '',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    last_active_at DATETIME NOT NULL,
    revoked_at DATETIME NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_sessions_user_id (user_id),
    INDEX idx_user_sessions_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Cot danh dau tai khoan tu vo hieu hoa (status = suspended + cot nay)
ALTER TABLE users
    ADD COLUMN self_deactivated_at DATETIME NULL AFTER email_verified_at;

-- 4. Cot ho tro doi email cho email_verification_tokens
ALTER TABLE email_verification_tokens
    ADD COLUMN purpose VARCHAR(20) NOT NULL DEFAULT 'verify_email' AFTER token,
    ADD COLUMN pending_email VARCHAR(255) NULL AFTER purpose;

-- Verify
-- SHOW COLUMNS FROM user_settings;
-- SHOW COLUMNS FROM user_sessions;
-- SHOW COLUMNS FROM users LIKE 'self_deactivated_at';
-- SHOW COLUMNS FROM email_verification_tokens LIKE 'purpose';
