-- Infrastructure Management Portal - Initial Schema Migration
-- This migration creates the core database structure for the application

-- Note: uuid-ossp extension has compatibility issues with Supabase PostgreSQL hooks
-- Using pgcrypto's gen_random_uuid() instead

-- Create custom schema for application
CREATE SCHEMA IF NOT EXISTS imp;

-- =====================================================
-- ROLES AND PERMISSIONS SYSTEM
-- =====================================================

-- Roles table
CREATE TABLE imp.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create default roles
INSERT INTO imp.roles (name, description, is_system_role) VALUES
    ('admin', 'Full read and write permissions including delete', TRUE),
    ('maintainer', 'Read and write permissions, but cannot delete', TRUE),
    ('viewer', 'Read-only permissions', TRUE);

-- User profiles table (extends Supabase auth.users)
CREATE TABLE imp.user_profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES imp.roles(id),
    full_name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id)
);

-- =====================================================
-- DYNAMIC SCHEMA MANAGEMENT
-- =====================================================

-- Data models table (for dynamic schema)
CREATE TABLE imp.data_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    table_name VARCHAR(100) UNIQUE NOT NULL,
    icon VARCHAR(50),
    is_system_model BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id)
);

-- Field types lookup
CREATE TABLE imp.field_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    sql_type VARCHAR(100) NOT NULL,
    validation_rules JSONB,
    ui_component VARCHAR(100),
    description TEXT
);

-- Insert standard field types
INSERT INTO imp.field_types (name, display_name, sql_type, ui_component) VALUES
    ('text', 'Text', 'VARCHAR(255)', 'text_input'),
    ('text_long', 'Long Text', 'TEXT', 'textarea'),
    ('number', 'Number', 'INTEGER', 'number_input'),
    ('decimal', 'Decimal', 'DECIMAL(10,2)', 'number_input'),
    ('boolean', 'Boolean', 'BOOLEAN', 'checkbox'),
    ('date', 'Date', 'DATE', 'date_picker'),
    ('datetime', 'Date Time', 'TIMESTAMP WITH TIME ZONE', 'datetime_picker'),
    ('email', 'Email', 'VARCHAR(255)', 'email_input'),
    ('url', 'URL', 'VARCHAR(500)', 'url_input'),
    ('json', 'JSON', 'JSONB', 'json_editor'),
    ('uuid', 'UUID', 'UUID', 'text_input'),
    ('reference', 'Reference', 'UUID', 'reference_select');

-- Fields table (for dynamic fields)
CREATE TABLE imp.fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES imp.data_models(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    field_type_id UUID NOT NULL REFERENCES imp.field_types(id),
    is_required BOOLEAN DEFAULT FALSE,
    is_unique BOOLEAN DEFAULT FALSE,
    default_value TEXT,
    validation_rules JSONB,
    reference_model_id UUID REFERENCES imp.data_models(id),
    display_order INTEGER DEFAULT 0,
    description TEXT,
    is_system_field BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id),
    UNIQUE (model_id, name)
);

-- Permissions table (role-based access per model)
CREATE TABLE imp.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES imp.roles(id) ON DELETE CASCADE,
    model_id UUID NOT NULL REFERENCES imp.data_models(id) ON DELETE CASCADE,
    can_read BOOLEAN DEFAULT FALSE,
    can_create BOOLEAN DEFAULT FALSE,
    can_update BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (role_id, model_id)
);

-- =====================================================
-- AUDIT LOGGING
-- =====================================================

CREATE TABLE imp.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id),
    user_email VARCHAR(255),
    action VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, LOGIN, etc.
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for audit log queries
CREATE INDEX idx_audit_logs_user_id ON imp.audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON imp.audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON imp.audit_logs(created_at DESC);

-- =====================================================
-- CORE DATA MODELS (Servers, SSL Certs, Apps, Services)
-- =====================================================

-- Servers table
CREATE TABLE imp.servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    ram_gb INTEGER,
    cpu_cores INTEGER,
    cpu_model VARCHAR(255),
    description TEXT,
    group_name VARCHAR(100),
    location VARCHAR(100),
    os_name VARCHAR(100),
    os_version VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id)
);

-- SSL Certificates table
CREATE TABLE imp.ssl_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    common_name VARCHAR(255) NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    sans TEXT[], -- Subject Alternative Names as array
    issuer VARCHAR(255),
    serial_number VARCHAR(255),
    certificate_type VARCHAR(50), -- e.g., DV, OV, EV
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id)
);

-- Applications table
CREATE TABLE imp.applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    group_name VARCHAR(100),
    version VARCHAR(100),
    description TEXT,
    repository_url VARCHAR(500),
    documentation_url VARCHAR(500),
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id)
);

-- Services table
CREATE TABLE imp.services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    group_name VARCHAR(100),
    server_id UUID REFERENCES imp.servers(id) ON DELETE RESTRICT,
    application_id UUID REFERENCES imp.applications(id) ON DELETE SET NULL,
    fqdn VARCHAR(255),
    ip_address INET,
    port INTEGER,
    ssl_certificate_id UUID REFERENCES imp.ssl_certificates(id) ON DELETE SET NULL,
    protocol VARCHAR(50), -- e.g., HTTP, HTTPS, TCP, UDP
    status VARCHAR(50) DEFAULT 'active',
    health_check_url VARCHAR(500),
    description TEXT,
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id)
);

-- Create indexes for better query performance
CREATE INDEX idx_servers_hostname ON imp.servers(hostname);
CREATE INDEX idx_servers_ip_address ON imp.servers(ip_address);
CREATE INDEX idx_servers_group ON imp.servers(group_name);
CREATE INDEX idx_ssl_certificates_common_name ON imp.ssl_certificates(common_name);
CREATE INDEX idx_ssl_certificates_valid_to ON imp.ssl_certificates(valid_to);
CREATE INDEX idx_applications_name ON imp.applications(name);
CREATE INDEX idx_services_server_id ON imp.services(server_id);
CREATE INDEX idx_services_application_id ON imp.services(application_id);
CREATE INDEX idx_services_fqdn ON imp.services(fqdn);

-- =====================================================
-- REGISTER CORE MODELS IN DATA_MODELS TABLE
-- =====================================================

INSERT INTO imp.data_models (name, display_name, description, table_name, icon, is_system_model) VALUES
    ('servers', 'Servers', 'Physical and virtual servers', 'imp.servers', 'server', TRUE),
    ('ssl_certificates', 'SSL Certificates', 'SSL/TLS certificates', 'imp.ssl_certificates', 'shield', TRUE),
    ('applications', 'Applications', 'Software applications', 'imp.applications', 'app-window', TRUE),
    ('services', 'Services', 'Services running on servers', 'imp.services', 'network', TRUE);

-- =====================================================
-- FUNCTIONS AND TRIGGERS
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION imp.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at on all tables
CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON imp.user_profiles
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_data_models_updated_at BEFORE UPDATE ON imp.data_models
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_fields_updated_at BEFORE UPDATE ON imp.fields
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_servers_updated_at BEFORE UPDATE ON imp.servers
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_ssl_certificates_updated_at BEFORE UPDATE ON imp.ssl_certificates
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON imp.applications
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_services_updated_at BEFORE UPDATE ON imp.services
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON imp.roles
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON imp.permissions
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();

-- Function to log changes (audit trail)
CREATE OR REPLACE FUNCTION imp.log_audit_trail()
RETURNS TRIGGER AS $$
DECLARE
    user_email_val VARCHAR(255);
BEGIN
    -- Get user email if user_id is available
    IF (TG_OP = 'DELETE') THEN
        SELECT email INTO user_email_val FROM auth.users WHERE id = OLD.updated_by;
        INSERT INTO imp.audit_logs (user_id, user_email, action, entity_type, entity_id, old_values)
        VALUES (OLD.updated_by, user_email_val, TG_OP, TG_TABLE_NAME, OLD.id, row_to_json(OLD));
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        SELECT email INTO user_email_val FROM auth.users WHERE id = NEW.updated_by;
        INSERT INTO imp.audit_logs (user_id, user_email, action, entity_type, entity_id, old_values, new_values)
        VALUES (NEW.updated_by, user_email_val, TG_OP, TG_TABLE_NAME, NEW.id, row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        SELECT email INTO user_email_val FROM auth.users WHERE id = NEW.created_by;
        INSERT INTO imp.audit_logs (user_id, user_email, action, entity_type, entity_id, new_values)
        VALUES (NEW.created_by, user_email_val, TG_OP, TG_TABLE_NAME, NEW.id, row_to_json(NEW));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create audit triggers for core tables
CREATE TRIGGER audit_servers AFTER INSERT OR UPDATE OR DELETE ON imp.servers
    FOR EACH ROW EXECUTE FUNCTION imp.log_audit_trail();
CREATE TRIGGER audit_ssl_certificates AFTER INSERT OR UPDATE OR DELETE ON imp.ssl_certificates
    FOR EACH ROW EXECUTE FUNCTION imp.log_audit_trail();
CREATE TRIGGER audit_applications AFTER INSERT OR UPDATE OR DELETE ON imp.applications
    FOR EACH ROW EXECUTE FUNCTION imp.log_audit_trail();
CREATE TRIGGER audit_services AFTER INSERT OR UPDATE OR DELETE ON imp.services
    FOR EACH ROW EXECUTE FUNCTION imp.log_audit_trail();
CREATE TRIGGER audit_user_profiles AFTER INSERT OR UPDATE OR DELETE ON imp.user_profiles
    FOR EACH ROW EXECUTE FUNCTION imp.log_audit_trail();

-- Function to prevent deletion of servers with active services
CREATE OR REPLACE FUNCTION imp.prevent_server_deletion_with_services()
RETURNS TRIGGER AS $$
DECLARE
    service_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO service_count FROM imp.services WHERE server_id = OLD.id;
    IF service_count > 0 THEN
        RAISE EXCEPTION 'Cannot delete server with active services. Please remove or reassign services first.';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_server_deletion BEFORE DELETE ON imp.servers
    FOR EACH ROW EXECUTE FUNCTION imp.prevent_server_deletion_with_services();

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================

-- Enable RLS on all tables
ALTER TABLE imp.user_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.data_models ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.fields ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.field_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.servers ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.ssl_certificates ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.applications ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.services ENABLE ROW LEVEL SECURITY;
ALTER TABLE imp.audit_logs ENABLE ROW LEVEL SECURITY;

-- Helper function to get user's role
CREATE OR REPLACE FUNCTION imp.get_user_role()
RETURNS VARCHAR AS $$
    SELECT r.name
    FROM imp.user_profiles up
    JOIN imp.roles r ON r.id = up.role_id
    WHERE up.id = auth.uid();
$$ LANGUAGE sql SECURITY DEFINER;

-- Helper function to check if user has permission
CREATE OR REPLACE FUNCTION imp.has_permission(model_name TEXT, permission_type TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    has_perm BOOLEAN;
    role_name VARCHAR;
BEGIN
    SELECT imp.get_user_role() INTO role_name;
    
    -- Admin has all permissions
    IF role_name = 'admin' THEN
        RETURN TRUE;
    END IF;
    
    -- Check specific permission
    SELECT 
        CASE permission_type
            WHEN 'read' THEN p.can_read
            WHEN 'create' THEN p.can_create
            WHEN 'update' THEN p.can_update
            WHEN 'delete' THEN p.can_delete
            ELSE FALSE
        END INTO has_perm
    FROM imp.permissions p
    JOIN imp.roles r ON r.id = p.role_id
    JOIN imp.data_models dm ON dm.id = p.model_id
    WHERE r.name = role_name AND dm.table_name = model_name;
    
    RETURN COALESCE(has_perm, FALSE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- RLS Policies for servers
CREATE POLICY "Users can view servers if they have read permission"
    ON imp.servers FOR SELECT
    USING (imp.has_permission('imp.servers', 'read'));

CREATE POLICY "Users can create servers if they have create permission"
    ON imp.servers FOR INSERT
    WITH CHECK (imp.has_permission('imp.servers', 'create'));

CREATE POLICY "Users can update servers if they have update permission"
    ON imp.servers FOR UPDATE
    USING (imp.has_permission('imp.servers', 'update'));

CREATE POLICY "Users can delete servers if they have delete permission"
    ON imp.servers FOR DELETE
    USING (imp.has_permission('imp.servers', 'delete'));

-- RLS Policies for SSL certificates (same pattern)
CREATE POLICY "Users can view ssl_certificates if they have read permission"
    ON imp.ssl_certificates FOR SELECT
    USING (imp.has_permission('imp.ssl_certificates', 'read'));

CREATE POLICY "Users can create ssl_certificates if they have create permission"
    ON imp.ssl_certificates FOR INSERT
    WITH CHECK (imp.has_permission('imp.ssl_certificates', 'create'));

CREATE POLICY "Users can update ssl_certificates if they have update permission"
    ON imp.ssl_certificates FOR UPDATE
    USING (imp.has_permission('imp.ssl_certificates', 'update'));

CREATE POLICY "Users can delete ssl_certificates if they have delete permission"
    ON imp.ssl_certificates FOR DELETE
    USING (imp.has_permission('imp.ssl_certificates', 'delete'));

-- RLS Policies for applications
CREATE POLICY "Users can view applications if they have read permission"
    ON imp.applications FOR SELECT
    USING (imp.has_permission('imp.applications', 'read'));

CREATE POLICY "Users can create applications if they have create permission"
    ON imp.applications FOR INSERT
    WITH CHECK (imp.has_permission('imp.applications', 'create'));

CREATE POLICY "Users can update applications if they have update permission"
    ON imp.applications FOR UPDATE
    USING (imp.has_permission('imp.applications', 'update'));

CREATE POLICY "Users can delete applications if they have delete permission"
    ON imp.applications FOR DELETE
    USING (imp.has_permission('imp.applications', 'delete'));

-- RLS Policies for services
CREATE POLICY "Users can view services if they have read permission"
    ON imp.services FOR SELECT
    USING (imp.has_permission('imp.services', 'read'));

CREATE POLICY "Users can create services if they have create permission"
    ON imp.services FOR INSERT
    WITH CHECK (imp.has_permission('imp.services', 'create'));

CREATE POLICY "Users can update services if they have update permission"
    ON imp.services FOR UPDATE
    USING (imp.has_permission('imp.services', 'update'));

CREATE POLICY "Users can delete services if they have delete permission"
    ON imp.services FOR DELETE
    USING (imp.has_permission('imp.services', 'delete'));

-- RLS Policies for audit logs (read-only for admins)
CREATE POLICY "Only admins can view audit logs"
    ON imp.audit_logs FOR SELECT
    USING (imp.get_user_role() = 'admin');

-- RLS Policies for user profiles
CREATE POLICY "Admins can view all user profiles"
    ON imp.user_profiles FOR SELECT
    USING (imp.get_user_role() = 'admin');

CREATE POLICY "Users can view their own profile"
    ON imp.user_profiles FOR SELECT
    USING (id = auth.uid());

CREATE POLICY "Admins can manage user profiles"
    ON imp.user_profiles FOR ALL
    USING (imp.get_user_role() = 'admin');

-- RLS Policies for roles (read-only for all authenticated users)
CREATE POLICY "Authenticated users can view roles"
    ON imp.roles FOR SELECT
    TO authenticated
    USING (true);

-- RLS Policies for data_models
CREATE POLICY "Authenticated users can view data models"
    ON imp.data_models FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Admins can manage data models"
    ON imp.data_models FOR ALL
    USING (imp.get_user_role() = 'admin');

-- RLS Policies for fields
CREATE POLICY "Authenticated users can view fields"
    ON imp.fields FOR SELECT
    TO authenticated
    USING (true);

CREATE POLICY "Admins can manage fields"
    ON imp.fields FOR ALL
    USING (imp.get_user_role() = 'admin');

-- RLS Policies for field_types
CREATE POLICY "Authenticated users can view field types"
    ON imp.field_types FOR SELECT
    TO authenticated
    USING (true);

-- RLS Policies for permissions
CREATE POLICY "Admins can manage permissions"
    ON imp.permissions FOR ALL
    USING (imp.get_user_role() = 'admin');

CREATE POLICY "Users can view their role permissions"
    ON imp.permissions FOR SELECT
    USING (
        role_id IN (
            SELECT role_id FROM imp.user_profiles WHERE id = auth.uid()
        )
    );

-- =====================================================
-- GRANT PERMISSIONS TO AUTHENTICATED ROLE
-- =====================================================

GRANT USAGE ON SCHEMA imp TO authenticated;
GRANT ALL ON ALL TABLES IN SCHEMA imp TO authenticated;
GRANT ALL ON ALL SEQUENCES IN SCHEMA imp TO authenticated;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA imp TO authenticated;

-- Grant to service_role for admin operations
GRANT USAGE ON SCHEMA imp TO service_role;
GRANT ALL ON ALL TABLES IN SCHEMA imp TO service_role;
GRANT ALL ON ALL SEQUENCES IN SCHEMA imp TO service_role;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA imp TO service_role;

-- =====================================================
-- INITIAL DATA SETUP
-- =====================================================

-- Set default permissions for roles
DO $$
DECLARE
    admin_role_id UUID;
    maintainer_role_id UUID;
    viewer_role_id UUID;
    model_record RECORD;
BEGIN
    -- Get role IDs
    SELECT id INTO admin_role_id FROM imp.roles WHERE name = 'admin';
    SELECT id INTO maintainer_role_id FROM imp.roles WHERE name = 'maintainer';
    SELECT id INTO viewer_role_id FROM imp.roles WHERE name = 'viewer';
    
    -- Set permissions for each model
    FOR model_record IN SELECT id FROM imp.data_models LOOP
        -- Admin: full access
        INSERT INTO imp.permissions (role_id, model_id, can_read, can_create, can_update, can_delete)
        VALUES (admin_role_id, model_record.id, TRUE, TRUE, TRUE, TRUE);
        
        -- Maintainer: read, create, update (no delete)
        INSERT INTO imp.permissions (role_id, model_id, can_read, can_create, can_update, can_delete)
        VALUES (maintainer_role_id, model_record.id, TRUE, TRUE, TRUE, FALSE);
        
        -- Viewer: read only
        INSERT INTO imp.permissions (role_id, model_id, can_read, can_create, can_update, can_delete)
        VALUES (viewer_role_id, model_record.id, TRUE, FALSE, FALSE, FALSE);
    END LOOP;
END $$;

-- Migration complete
SELECT 'Infrastructure Management Portal schema initialized successfully!' AS status;
