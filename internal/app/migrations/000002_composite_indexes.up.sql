CREATE INDEX IF NOT EXISTS idx_intake_leads_client_internal_created ON intake_leads(client_status, internal_status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_contact_messages_status_created ON contact_messages(status, created_at DESC);
