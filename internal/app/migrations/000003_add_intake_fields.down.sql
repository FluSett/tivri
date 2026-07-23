DROP INDEX IF EXISTS idx_intake_leads_service_type;

ALTER TABLE intake_leads DROP COLUMN IF EXISTS service_type;
ALTER TABLE intake_leads DROP COLUMN IF EXISTS existing_url;
ALTER TABLE intake_leads DROP COLUMN IF EXISTS tech_stack;
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS is_custom_budget BOOLEAN NOT NULL DEFAULT FALSE;
