-- Rollback: Remove DC info columns from accounts table

ALTER TABLE accounts DROP COLUMN IF EXISTS dc_id;
ALTER TABLE accounts DROP COLUMN IF EXISTS dc_addr;
