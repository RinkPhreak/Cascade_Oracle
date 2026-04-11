-- Add DC info columns to accounts table for persisting Telegram Data Center routing info
-- This ensures correct DC routing on reconnect without re-parsing session files

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS dc_id INTEGER DEFAULT NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS dc_addr VARCHAR(255) DEFAULT NULL;
