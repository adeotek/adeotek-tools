-- Seed file for default admin user
-- Note: In production, you would create users through Supabase Auth API
-- This is for development/testing purposes only

-- Insert default admin user into auth.users (if using local Supabase)
-- Password: Admin@123456 (CHANGE THIS IN PRODUCTION!)
-- This will be handled by the application on first run

-- Insert default entities for demonstration
INSERT INTO entities (name, display_name, description, icon) VALUES
    ('server', 'Server', 'Physical or virtual servers', 'server'),
    ('ssl_certificate', 'SSL Certificate', 'SSL/TLS certificates', 'shield-check'),
    ('database', 'Database', 'Database instances', 'database'),
    ('application', 'Application', 'Applications and services', 'app-window')
ON CONFLICT (name) DO NOTHING;

-- Insert default attributes for Server entity
DO $$
DECLARE
    server_entity_id UUID;
BEGIN
    SELECT id INTO server_entity_id FROM entities WHERE name = 'server';
    
    INSERT INTO attributes (entity_id, name, display_name, attribute_type, is_required, order_index) VALUES
        (server_entity_id, 'hostname', 'Hostname', 'string', TRUE, 1),
        (server_entity_id, 'ip_address', 'IP Address', 'string', TRUE, 2),
        (server_entity_id, 'os', 'Operating System', 'string', FALSE, 3),
        (server_entity_id, 'cpu_cores', 'CPU Cores', 'number', FALSE, 4),
        (server_entity_id, 'ram_gb', 'RAM (GB)', 'number', FALSE, 5),
        (server_entity_id, 'purchase_date', 'Purchase Date', 'date', FALSE, 6),
        (server_entity_id, 'notes', 'Notes', 'string', FALSE, 7)
    ON CONFLICT (entity_id, name) DO NOTHING;
END $$;

-- Insert default attributes for SSL Certificate entity
DO $$
DECLARE
    ssl_entity_id UUID;
    server_entity_id UUID;
BEGIN
    SELECT id INTO ssl_entity_id FROM entities WHERE name = 'ssl_certificate';
    SELECT id INTO server_entity_id FROM entities WHERE name = 'server';
    
    INSERT INTO attributes (entity_id, name, display_name, attribute_type, is_required, references_entity_id, order_index) VALUES
        (ssl_entity_id, 'domain', 'Domain', 'string', TRUE, NULL, 1),
        (ssl_entity_id, 'issuer', 'Issuer', 'string', FALSE, NULL, 2),
        (ssl_entity_id, 'issue_date', 'Issue Date', 'date', FALSE, NULL, 3),
        (ssl_entity_id, 'expiry_date', 'Expiry Date', 'date', TRUE, NULL, 4),
        (ssl_entity_id, 'server', 'Server', 'relation', FALSE, server_entity_id, 5)
    ON CONFLICT (entity_id, name) DO NOTHING;
END $$;

-- Insert default attributes for Database entity
DO $$
DECLARE
    database_entity_id UUID;
    server_entity_id UUID;
BEGIN
    SELECT id INTO database_entity_id FROM entities WHERE name = 'database';
    SELECT id INTO server_entity_id FROM entities WHERE name = 'server';
    
    INSERT INTO attributes (entity_id, name, display_name, attribute_type, is_required, references_entity_id, order_index) VALUES
        (database_entity_id, 'name', 'Database Name', 'string', TRUE, NULL, 1),
        (database_entity_id, 'type', 'Database Type', 'string', TRUE, NULL, 2),
        (database_entity_id, 'version', 'Version', 'string', FALSE, NULL, 3),
        (database_entity_id, 'size_gb', 'Size (GB)', 'number', FALSE, NULL, 4),
        (database_entity_id, 'server', 'Server', 'relation', TRUE, server_entity_id, 5)
    ON CONFLICT (entity_id, name) DO NOTHING;
END $$;

-- Insert default attributes for Application entity
DO $$
DECLARE
    app_entity_id UUID;
    server_entity_id UUID;
    database_entity_id UUID;
BEGIN
    SELECT id INTO app_entity_id FROM entities WHERE name = 'application';
    SELECT id INTO server_entity_id FROM entities WHERE name = 'server';
    SELECT id INTO database_entity_id FROM entities WHERE name = 'database';
    
    INSERT INTO attributes (entity_id, name, display_name, attribute_type, is_required, references_entity_id, order_index) VALUES
        (app_entity_id, 'name', 'Application Name', 'string', TRUE, NULL, 1),
        (app_entity_id, 'version', 'Version', 'string', FALSE, NULL, 2),
        (app_entity_id, 'url', 'URL', 'string', FALSE, NULL, 3),
        (app_entity_id, 'server', 'Server', 'relation', FALSE, server_entity_id, 4),
        (app_entity_id, 'database', 'Database', 'relation', FALSE, database_entity_id, 5)
    ON CONFLICT (entity_id, name) DO NOTHING;
END $$;
