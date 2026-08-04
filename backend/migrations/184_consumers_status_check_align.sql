-- 184: Align consumers.status CHECK constraint with Ent schema / service code
-- Migration 176 only allowed ('active', 'suspended', 'deleted'),
-- but the service layer and frontend use 'inactive' (not 'suspended'/'deleted').
-- Soft delete is handled via deleted_at, so 'deleted' is not a valid status value.

ALTER TABLE consumers DROP CONSTRAINT IF EXISTS consumers_status_check;
ALTER TABLE consumers ADD CONSTRAINT consumers_status_check
    CHECK (status IN ('active', 'inactive', 'suspended'));
