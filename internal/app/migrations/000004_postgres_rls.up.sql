-- Migration 000004: Enable PostgreSQL RLS and GIN Trigram Search Indexes

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 1. Enable and Force RLS on all system tables
ALTER TABLE intake_leads ENABLE ROW LEVEL SECURITY;
ALTER TABLE intake_leads FORCE ROW LEVEL SECURITY;

ALTER TABLE contact_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE contact_messages FORCE ROW LEVEL SECURITY;

ALTER TABLE admin_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE admin_sessions FORCE ROW LEVEL SECURITY;

ALTER TABLE system_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE system_settings FORCE ROW LEVEL SECURITY;

ALTER TABLE portfolio_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE portfolio_items FORCE ROW LEVEL SECURITY;

ALTER TABLE outbox_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE outbox_events FORCE ROW LEVEL SECURITY;

-- 2. Intake Leads Policies
-- Public can submit new leads (INSERT)
CREATE POLICY rls_intake_leads_public_insert ON intake_leads
    FOR INSERT
    WITH CHECK (true);

-- Admin and System can SELECT, UPDATE, DELETE
CREATE POLICY rls_intake_leads_admin_all ON intake_leads
    FOR ALL
    USING ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'))
    WITH CHECK ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'));

-- 3. Contact Messages Policies
CREATE POLICY rls_contact_messages_public_insert ON contact_messages
    FOR INSERT
    WITH CHECK (true);

CREATE POLICY rls_contact_messages_admin_all ON contact_messages
    FOR ALL
    USING ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'))
    WITH CHECK ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'));

-- 4. Portfolio Items Policies (Public READ, Admin MUTATE)
CREATE POLICY rls_portfolio_items_read ON portfolio_items
    FOR SELECT
    USING (true);

CREATE POLICY rls_portfolio_items_admin_mutate ON portfolio_items
    FOR ALL
    USING ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'))
    WITH CHECK ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'));

-- 5. System Settings Policies (Public READ, Admin MUTATE)
CREATE POLICY rls_system_settings_read ON system_settings
    FOR SELECT
    USING (true);

CREATE POLICY rls_system_settings_admin_mutate ON system_settings
    FOR ALL
    USING ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'))
    WITH CHECK ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'));

-- 6. Admin Sessions Policies
CREATE POLICY rls_admin_sessions_all ON admin_sessions
    FOR ALL
    USING ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'))
    WITH CHECK ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'));

-- 7. Outbox Events Policies
CREATE POLICY rls_outbox_events_insert ON outbox_events
    FOR INSERT
    WITH CHECK (true);

CREATE POLICY rls_outbox_events_admin_all ON outbox_events
    FOR ALL
    USING ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'))
    WITH CHECK ((SELECT current_setting('app.current_role', true)) IN ('admin', 'system'));

-- 8. High-Load Trigram GIN and Covering Composite Indexes
CREATE INDEX IF NOT EXISTS idx_intake_leads_company_trgm ON intake_leads USING gin (company_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_intake_leads_email_trgm ON intake_leads USING gin (contact_email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_contact_messages_email_trgm ON contact_messages USING gin (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_contact_messages_topic_trgm ON contact_messages USING gin (topic gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_intake_leads_status_created_covering ON intake_leads(client_status, internal_status, created_at DESC) INCLUDE (company_name, contact_email, budget);
