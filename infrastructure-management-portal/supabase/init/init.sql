-- Initialize database permissions for Supabase services
-- This script runs during PostgreSQL initialization via docker-entrypoint-initdb.d

-- Grant permissions on public schema to Supabase auth admin role
GRANT CREATE ON SCHEMA public TO supabase_auth_admin;
GRANT ALL PRIVILEGES ON SCHEMA public TO supabase_auth_admin;

-- Grant permissions to other Supabase service roles
GRANT CREATE ON SCHEMA public TO supabase_storage_admin;
GRANT ALL PRIVILEGES ON SCHEMA public TO supabase_storage_admin;

-- Grant permissions to admin role
GRANT ALL PRIVILEGES ON SCHEMA public TO supabase_admin;

-- Grant usage to authenticated and service roles
GRANT USAGE ON SCHEMA public TO authenticated;
GRANT USAGE ON SCHEMA public TO service_role;
GRANT USAGE ON SCHEMA public TO anon;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO supabase_auth_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO supabase_storage_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO supabase_auth_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO supabase_storage_admin;
