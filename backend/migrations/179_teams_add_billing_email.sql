-- 179: Add billing_email column to teams table
-- Ent schema defines billing_email field but database column was missing
ALTER TABLE teams ADD COLUMN IF NOT EXISTS billing_email VARCHAR(255);
