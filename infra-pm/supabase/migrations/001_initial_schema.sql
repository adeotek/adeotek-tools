-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create custom types
CREATE TYPE user_role AS ENUM ('admin', 'maintainer', 'viewer');
CREATE TYPE attribute_type AS ENUM ('string', 'number', 'date', 'relation');
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete', 'view');

-- =============================================
-- RBAC Tables
-- =============================================

-- Profiles table (extends auth.users)
CREATE TABLE profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    email TEXT NOT NULL UNIQUE,
    full_name TEXT,
    avatar_url TEXT,
    role user_role NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on role for faster lookups
CREATE INDEX idx_profiles_role ON profiles(role);

-- Roles table (for granular permissions)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'Full access to all resources'),
    ('maintainer', 'Read and write access, cannot delete'),
    ('viewer', 'Read-only access');

-- Permissions table (entity-level permissions)
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    entity_type_id UUID, -- NULL means applies to all entities
    can_create BOOLEAN DEFAULT FALSE,
    can_read BOOLEAN DEFAULT TRUE,
    can_update BOOLEAN DEFAULT FALSE,
    can_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, entity_type_id)
);

-- Insert default permissions
-- Admin: full access
INSERT INTO permissions (role_id, entity_type_id, can_create, can_read, can_update, can_delete)
SELECT id, NULL, TRUE, TRUE, TRUE, TRUE FROM roles WHERE name = 'admin';

-- Maintainer: read/write, no delete
INSERT INTO permissions (role_id, entity_type_id, can_create, can_read, can_update, can_delete)
SELECT id, NULL, TRUE, TRUE, TRUE, FALSE FROM roles WHERE name = 'maintainer';

-- Viewer: read-only
INSERT INTO permissions (role_id, entity_type_id, can_create, can_read, can_update, can_delete)
SELECT id, NULL, FALSE, TRUE, FALSE, FALSE FROM roles WHERE name = 'viewer';

-- =============================================
-- Dynamic Schema Tables
-- =============================================

-- Entities table (defines types like "Server", "SSL Cert")
CREATE TABLE entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    icon TEXT, -- lucide icon name
    created_by UUID REFERENCES auth.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX idx_entities_name ON entities(name);

-- Attributes table (defines fields for entities)
CREATE TABLE attributes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    attribute_type attribute_type NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    default_value TEXT,
    references_entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    validation_rules JSONB, -- For additional validation (min, max, pattern, etc.)
    order_index INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(entity_id, name),
    CHECK (
        (attribute_type = 'relation' AND references_entity_id IS NOT NULL) OR
        (attribute_type != 'relation' AND references_entity_id IS NULL)
    )
);

-- Create indexes
CREATE INDEX idx_attributes_entity_id ON attributes(entity_id);
CREATE INDEX idx_attributes_references_entity_id ON attributes(references_entity_id);

-- Records table (stores instances of entities)
CREATE TABLE records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    created_by UUID REFERENCES auth.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX idx_records_entity_id ON records(entity_id);
CREATE INDEX idx_records_created_by ON records(created_by);

-- Values table (stores data for each attribute per record)
CREATE TABLE values (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    record_id UUID NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    attribute_id UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    value_text TEXT,
    value_number NUMERIC,
    value_date TIMESTAMP WITH TIME ZONE,
    value_relation UUID REFERENCES records(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(record_id, attribute_id)
);

-- Create indexes
CREATE INDEX idx_values_record_id ON values(record_id);
CREATE INDEX idx_values_attribute_id ON values(attribute_id);
CREATE INDEX idx_values_relation ON values(value_relation) WHERE value_relation IS NOT NULL;

-- =============================================
-- Audit Log Table
-- =============================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    action audit_action NOT NULL,
    entity_type TEXT NOT NULL,
    record_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_record_id ON audit_logs(record_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- =============================================
-- Functions
-- =============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to check if a record can be deleted (no references)
CREATE OR REPLACE FUNCTION check_record_references(record_uuid UUID)
RETURNS BOOLEAN AS $$
DECLARE
    ref_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO ref_count
    FROM values
    WHERE value_relation = record_uuid;
    
    RETURN ref_count = 0;
END;
$$ LANGUAGE plpgsql;

-- Function to log audit events
CREATE OR REPLACE FUNCTION log_audit_event()
RETURNS TRIGGER AS $$
DECLARE
    action_type audit_action;
    entity_name TEXT;
BEGIN
    -- Determine action type
    IF TG_OP = 'INSERT' THEN
        action_type := 'create';
    ELSIF TG_OP = 'UPDATE' THEN
        action_type := 'update';
    ELSIF TG_OP = 'DELETE' THEN
        action_type := 'delete';
    END IF;

    -- Get entity name if applicable
    IF TG_TABLE_NAME = 'records' THEN
        IF TG_OP = 'DELETE' THEN
            SELECT name INTO entity_name FROM entities WHERE id = OLD.entity_id;
            INSERT INTO audit_logs (user_id, action, entity_type, record_id, old_values)
            VALUES (
                OLD.created_by,
                action_type,
                COALESCE(entity_name, 'unknown'),
                OLD.id,
                row_to_json(OLD)::jsonb
            );
        ELSE
            SELECT name INTO entity_name FROM entities WHERE id = NEW.entity_id;
            INSERT INTO audit_logs (user_id, action, entity_type, record_id, new_values, old_values)
            VALUES (
                NEW.created_by,
                action_type,
                COALESCE(entity_name, 'unknown'),
                NEW.id,
                row_to_json(NEW)::jsonb,
                CASE WHEN TG_OP = 'UPDATE' THEN row_to_json(OLD)::jsonb ELSE NULL END
            );
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- Triggers
-- =============================================

-- Triggers for updated_at
CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entities_updated_at BEFORE UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_attributes_updated_at BEFORE UPDATE ON attributes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_records_updated_at BEFORE UPDATE ON records
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_values_updated_at BEFORE UPDATE ON values
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Audit triggers
CREATE TRIGGER audit_records_changes
    AFTER INSERT OR UPDATE OR DELETE ON records
    FOR EACH ROW EXECUTE FUNCTION log_audit_event();

-- =============================================
-- Row Level Security (RLS) Policies
-- =============================================

-- Enable RLS on all tables
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE attributes ENABLE ROW LEVEL SECURITY;
ALTER TABLE records ENABLE ROW LEVEL SECURITY;
ALTER TABLE values ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- Profiles policies
CREATE POLICY "Users can view all profiles"
    ON profiles FOR SELECT
    USING (true);

CREATE POLICY "Users can update their own profile"
    ON profiles FOR UPDATE
    USING (auth.uid() = id);

CREATE POLICY "Admins can do everything on profiles"
    ON profiles FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role = 'admin'
        )
    );

-- Roles policies (read-only for non-admins)
CREATE POLICY "Everyone can view roles"
    ON roles FOR SELECT
    USING (true);

CREATE POLICY "Only admins can modify roles"
    ON roles FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role = 'admin'
        )
    );

-- Permissions policies
CREATE POLICY "Everyone can view permissions"
    ON permissions FOR SELECT
    USING (true);

CREATE POLICY "Only admins can modify permissions"
    ON permissions FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role = 'admin'
        )
    );

-- Entities policies
CREATE POLICY "Everyone can view entities"
    ON entities FOR SELECT
    USING (true);

CREATE POLICY "Admins can do everything on entities"
    ON entities FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role = 'admin'
        )
    );

CREATE POLICY "Maintainers can create and update entities"
    ON entities FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role IN ('admin', 'maintainer')
        )
    );

CREATE POLICY "Maintainers can update entities"
    ON entities FOR UPDATE
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role IN ('admin', 'maintainer')
        )
    );

-- Attributes policies
CREATE POLICY "Everyone can view attributes"
    ON attributes FOR SELECT
    USING (true);

CREATE POLICY "Admins and maintainers can manage attributes"
    ON attributes FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role IN ('admin', 'maintainer')
        )
    );

-- Records policies
CREATE POLICY "Everyone can view records"
    ON records FOR SELECT
    USING (true);

CREATE POLICY "Admins can do everything on records"
    ON records FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role = 'admin'
        )
    );

CREATE POLICY "Maintainers can create and update records"
    ON records FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role IN ('admin', 'maintainer')
        )
    );

CREATE POLICY "Maintainers can update records"
    ON records FOR UPDATE
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role IN ('admin', 'maintainer')
        )
    );

-- Values policies
CREATE POLICY "Everyone can view values"
    ON values FOR SELECT
    USING (true);

CREATE POLICY "Admins and maintainers can manage values"
    ON values FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM profiles
            WHERE id = auth.uid() AND role IN ('admin', 'maintainer')
        )
    );

-- Audit logs policies
CREATE POLICY "Everyone can view audit logs"
    ON audit_logs FOR SELECT
    USING (true);

CREATE POLICY "System can insert audit logs"
    ON audit_logs FOR INSERT
    WITH CHECK (true);

-- =============================================
-- Helper Functions for Application
-- =============================================

-- Function to get user role
CREATE OR REPLACE FUNCTION get_user_role(user_uuid UUID)
RETURNS user_role AS $$
DECLARE
    user_role_result user_role;
BEGIN
    SELECT role INTO user_role_result
    FROM profiles
    WHERE id = user_uuid;
    
    RETURN user_role_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check if user has permission
CREATE OR REPLACE FUNCTION has_permission(
    user_uuid UUID,
    entity_uuid UUID,
    permission_type TEXT
)
RETURNS BOOLEAN AS $$
DECLARE
    user_role_val user_role;
    role_uuid UUID;
    has_perm BOOLEAN;
BEGIN
    -- Get user role
    SELECT role INTO user_role_val FROM profiles WHERE id = user_uuid;
    
    -- Admin has all permissions
    IF user_role_val = 'admin' THEN
        RETURN TRUE;
    END IF;
    
    -- Get role id
    SELECT id INTO role_uuid FROM roles WHERE name = user_role_val::TEXT;
    
    -- Check permission
    IF permission_type = 'create' THEN
        SELECT can_create INTO has_perm FROM permissions
        WHERE role_id = role_uuid AND (entity_type_id = entity_uuid OR entity_type_id IS NULL)
        ORDER BY entity_type_id NULLS LAST LIMIT 1;
    ELSIF permission_type = 'read' THEN
        SELECT can_read INTO has_perm FROM permissions
        WHERE role_id = role_uuid AND (entity_type_id = entity_uuid OR entity_type_id IS NULL)
        ORDER BY entity_type_id NULLS LAST LIMIT 1;
    ELSIF permission_type = 'update' THEN
        SELECT can_update INTO has_perm FROM permissions
        WHERE role_id = role_uuid AND (entity_type_id = entity_uuid OR entity_type_id IS NULL)
        ORDER BY entity_type_id NULLS LAST LIMIT 1;
    ELSIF permission_type = 'delete' THEN
        SELECT can_delete INTO has_perm FROM permissions
        WHERE role_id = role_uuid AND (entity_type_id = entity_uuid OR entity_type_id IS NULL)
        ORDER BY entity_type_id NULLS LAST LIMIT 1;
    END IF;
    
    RETURN COALESCE(has_perm, FALSE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Grant necessary permissions
GRANT USAGE ON SCHEMA public TO anon, authenticated;
GRANT ALL ON ALL TABLES IN SCHEMA public TO anon, authenticated;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO anon, authenticated;
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO anon, authenticated;
