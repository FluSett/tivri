TRUNCATE TABLE intake_leads;

ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS service_type TEXT NOT NULL DEFAULT 'full_project';
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS existing_url TEXT NOT NULL DEFAULT '';
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS tech_stack TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name='intake_leads' AND column_name='is_custom_budget'
    ) THEN
        ALTER TABLE intake_leads DROP COLUMN is_custom_budget;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_intake_leads_service_type ON intake_leads(service_type);
