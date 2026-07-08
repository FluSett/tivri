CREATE TABLE IF NOT EXISTS portfolio_items (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    budget BIGINT NOT NULL,
    tech_stack TEXT NOT NULL,
    media TEXT NOT NULL DEFAULT '[]'
);
CREATE TABLE IF NOT EXISTS intake_leads (
    id SERIAL PRIMARY KEY,
    company_name TEXT NOT NULL,
    project_scope TEXT NOT NULL,
    budget BIGINT NOT NULL,
    contact_email TEXT NOT NULL DEFAULT '',
    contact_info TEXT NOT NULL DEFAULT '',
    client_status TEXT NOT NULL DEFAULT 'pending',
    internal_status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS contact_messages (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    topic TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'new',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE portfolio_items ADD COLUMN IF NOT EXISTS media TEXT NOT NULL DEFAULT '[]';
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS client_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS internal_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE contact_messages ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'new';
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
INSERT INTO system_settings (key, value) VALUES ('high_queue', 'false') ON CONFLICT (key) DO NOTHING;

ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS deadline_needed BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS deadline_spec TEXT NOT NULL DEFAULT '';
ALTER TABLE intake_leads ADD COLUMN IF NOT EXISTS is_custom_budget BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name='intake_leads' AND column_name='contact_phone'
    ) THEN
        ALTER TABLE intake_leads RENAME COLUMN contact_phone TO contact_info;
    END IF;
END $$;

INSERT INTO system_settings (key, value) VALUES ('maintenance_mode', 'false') ON CONFLICT (key) DO NOTHING;
