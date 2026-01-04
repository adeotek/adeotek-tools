#!/bin/bash
set -e

# This script creates the required Supabase roles before other initialization scripts run
# It must run before init.sql (which grants permissions to these roles)

echo "Creating Supabase roles..."

# Wait for PostgreSQL to be ready
until pg_isready -U postgres; do
  echo "Waiting for PostgreSQL..."
  sleep 1
done

# Create required Supabase roles
psql -v ON_ERROR_STOP=1 --username postgres --dbname postgres <<-EOSQL
  -- Create required roles if they don't exist
  DO \$\$
  BEGIN
    -- Supabase admin role
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'supabase_admin') THEN
      CREATE ROLE supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS PASSWORD '${POSTGRES_PASSWORD}';
      RAISE NOTICE 'Created role: supabase_admin';
    END IF;

    -- Auth admin role
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'supabase_auth_admin') THEN
      CREATE ROLE supabase_auth_admin NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD '${POSTGRES_PASSWORD}';
      RAISE NOTICE 'Created role: supabase_auth_admin';
    END IF;

    -- Storage admin role
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'supabase_storage_admin') THEN
      CREATE ROLE supabase_storage_admin NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD '${POSTGRES_PASSWORD}';
      RAISE NOTICE 'Created role: supabase_storage_admin';
    END IF;

    -- Authenticator role (used by PostgREST)
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'authenticator') THEN
      CREATE ROLE authenticator NOINHERIT LOGIN NOREPLICATION PASSWORD '${POSTGRES_PASSWORD}';
      RAISE NOTICE 'Created role: authenticator';
    END IF;

    -- Anonymous role
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'anon') THEN
      CREATE ROLE anon NOLOGIN NOINHERIT;
      RAISE NOTICE 'Created role: anon';
    END IF;

    -- Authenticated role
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'authenticated') THEN
      CREATE ROLE authenticated NOLOGIN NOINHERIT;
      RAISE NOTICE 'Created role: authenticated';
    END IF;

    -- Service role (for admin operations)
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'service_role') THEN
      CREATE ROLE service_role NOLOGIN NOINHERIT BYPASSRLS;
      RAISE NOTICE 'Created role: service_role';
    END IF;
  END
  \$\$;

  -- Create auth schema for Supabase auth service
  CREATE SCHEMA IF NOT EXISTS auth;
  GRANT USAGE ON SCHEMA auth TO supabase_auth_admin;
  GRANT ALL ON SCHEMA auth TO supabase_auth_admin;
  GRANT USAGE ON SCHEMA auth TO postgres;
  ALTER SCHEMA auth OWNER TO supabase_auth_admin;

  -- Create storage schema for Supabase storage service
  CREATE SCHEMA IF NOT EXISTS storage;
  GRANT USAGE ON SCHEMA storage TO supabase_storage_admin;
  GRANT ALL ON SCHEMA storage TO supabase_storage_admin;
  GRANT USAGE ON SCHEMA storage TO postgres;
  ALTER SCHEMA storage OWNER TO supabase_storage_admin;

  -- Grant database level permissions
  GRANT ALL PRIVILEGES ON DATABASE postgres TO supabase_storage_admin;
  GRANT ALL PRIVILEGES ON DATABASE postgres TO supabase_admin;

  -- Create _realtime schema for Supabase realtime service
  CREATE SCHEMA IF NOT EXISTS _realtime;
  GRANT USAGE ON SCHEMA _realtime TO supabase_admin;
  GRANT ALL ON SCHEMA _realtime TO supabase_admin;
  ALTER SCHEMA _realtime OWNER TO supabase_admin;

  -- Note: Extensions (pgcrypto, uuid-ossp) have compatibility issues with Supabase hooks
  -- pgcrypto is already available in Supabase image - use gen_random_uuid() instead of uuid_generate_v4()

  -- Grant necessary role memberships
  GRANT anon, authenticated, service_role TO authenticator;
  GRANT supabase_auth_admin TO postgres;
  GRANT supabase_storage_admin TO postgres;
  GRANT supabase_admin TO postgres;
EOSQL

echo "Supabase roles created successfully!"
