DROP INDEX IF EXISTS idx_contact_messages_status;
DROP INDEX IF EXISTS idx_intake_leads_internal_status;
DROP INDEX IF EXISTS idx_intake_leads_client_status;
DROP INDEX IF EXISTS idx_outbox_events_unprocessed;

DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS admin_sessions;
DROP TABLE IF EXISTS contact_messages;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS intake_leads;
DROP TABLE IF EXISTS portfolio_items;
