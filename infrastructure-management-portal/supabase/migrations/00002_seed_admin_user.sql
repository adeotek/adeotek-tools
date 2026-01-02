-- Seed default admin user
-- This script creates the initial admin user for the system
-- The credentials should be changed immediately after first login

DO $$
DECLARE
    admin_user_id UUID;
    admin_role_id UUID;
    admin_email VARCHAR(255) := COALESCE(current_setting('app.default_admin_email', true), 'admin@example.com');
    admin_password VARCHAR(255) := COALESCE(current_setting('app.default_admin_password', true), 'ChangeMe123!');
    admin_name VARCHAR(255) := COALESCE(current_setting('app.default_admin_full_name', true), 'System Administrator');
BEGIN
    -- Get admin role ID
    SELECT id INTO admin_role_id FROM imp.roles WHERE name = 'admin';
    
    -- Check if admin user already exists
    SELECT id INTO admin_user_id FROM auth.users WHERE email = admin_email LIMIT 1;
    
    IF admin_user_id IS NULL THEN
        -- Create admin user in auth.users
        -- Note: This is a simplified version. In production, you should use Supabase Admin API
        -- or a proper user creation method that handles password hashing correctly
        
        INSERT INTO auth.users (
            instance_id,
            id,
            aud,
            role,
            email,
            encrypted_password,
            email_confirmed_at,
            raw_app_meta_data,
            raw_user_meta_data,
            created_at,
            updated_at,
            confirmation_token,
            email_change,
            email_change_token_new,
            recovery_token
        ) VALUES (
            '00000000-0000-0000-0000-000000000000',
            uuid_generate_v4(),
            'authenticated',
            'authenticated',
            admin_email,
            crypt(admin_password, gen_salt('bf')),
            NOW(),
            jsonb_build_object('provider', 'email', 'providers', ARRAY['email']),
            jsonb_build_object('full_name', admin_name),
            NOW(),
            NOW(),
            '',
            '',
            '',
            ''
        )
        RETURNING id INTO admin_user_id;
        
        -- Create user profile
        INSERT INTO imp.user_profiles (id, role_id, full_name, email, is_active, created_at)
        VALUES (admin_user_id, admin_role_id, admin_name, admin_email, TRUE, NOW());
        
        RAISE NOTICE 'Default admin user created: %', admin_email;
        RAISE NOTICE 'Please change the password after first login!';
    ELSE
        RAISE NOTICE 'Admin user already exists with email: %', admin_email;
    END IF;
END $$;
