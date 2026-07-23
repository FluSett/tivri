-- Down Migration 000004: Remove RLS policies and GIN Trigram Search Indexes

DROP INDEX IF EXISTS idx_intake_leads_status_created_covering;
DROP INDEX IF EXISTS idx_contact_messages_topic_trgm;
DROP INDEX IF EXISTS idx_contact_messages_email_trgm;
DROP INDEX IF EXISTS idx_intake_leads_email_trgm;
DROP INDEX IF EXISTS idx_intake_leads_company_trgm;

DROP POLICY IF EXISTS rls_outbox_events_admin_all ON outbox_events;
DROP POLICY IF EXISTS rls_outbox_events_insert ON outbox_events;

DROP POLICY IF EXISTS rls_admin_sessions_all ON admin_sessions;

DROP POLICY IF EXISTS rls_system_settings_admin_mutate ON system_settings;
DROP POLICY IF EXISTS rls_system_settings_read ON system_settings;

DROP POLICY IF EXISTS rls_portfolio_items_admin_mutate ON portfolio_items;
DROP POLICY IF EXISTS rls_portfolio_items_read ON portfolio_items;

DROP POLICY IF EXISTS rls_contact_messages_admin_all ON contact_messages;
DROP POLICY IF EXISTS rls_contact_messages_public_insert ON contact_messages;

DROP POLICY IF EXISTS rls_intake_leads_admin_all ON intake_leads;
DROP POLICY IF EXISTS rls_intake_leads_public_insert ON intake_leads;

ALTER TABLE outbox_events DISABLE ROW LEVEL SECURITY;
ALTER TABLE portfolio_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE system_settings DISABLE ROW LEVEL SECURITY;
ALTER TABLE admin_sessions DISABLE ROW LEVEL SECURITY;
ALTER TABLE contact_messages DISABLE ROW LEVEL SECURITY;
ALTER TABLE intake_leads DISABLE ROW LEVEL SECURITY;

DROP EXTENSION IF EXISTS pg_trgm;
