-- Migration 001: Add video call columns to calls table
-- 
-- Thêm 2 cột video_enabled_caller và video_enabled_callee
-- để hỗ trợ tính năng bật/tắt video trong video call.
--
-- Cách chạy:
--   mysql -u <user> -p <db_name> < docs/migrations/001_video_call_columns.sql
-- Hoặc paste vào MySQL client / phpMyAdmin.

ALTER TABLE calls
  ADD COLUMN video_enabled_caller TINYINT(1) NOT NULL DEFAULT 0
    AFTER muted_callee,
  ADD COLUMN video_enabled_callee TINYINT(1) NOT NULL DEFAULT 0
    AFTER video_enabled_caller;

-- Verify
-- SELECT id, call_type, video_enabled_caller, video_enabled_callee FROM calls;
