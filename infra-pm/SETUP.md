# InfraPM - Quick Setup Guide

This guide will help you get InfraPM up and running in minutes.

## Prerequisites

- Node.js 20.x or later
- Docker and Docker Compose
- npm

## Step 1: Install Dependencies

```bash
cd infra-pm
npm install
```

## Step 2: Configure Environment

```bash
cp .env.example .env.local
```

**Important**: Update these values in `.env.local`:

```env
# Generate secure values for these:
POSTGRES_PASSWORD=your-secure-password-here
JWT_SECRET=your-jwt-secret-32-characters-minimum
SECRET_KEY_BASE=your-secret-key-base-64-characters-minimum
```

To generate secure secrets, you can use:
```bash
# For JWT_SECRET (32+ characters)
openssl rand -base64 32

# For SECRET_KEY_BASE (64+ characters)
openssl rand -base64 64
```

## Step 3: Start Supabase

```bash
docker-compose up -d
```

Wait a few moments for all services to start. You can check the status with:
```bash
docker-compose ps
```

## Step 4: Verify Database Migrations

The migrations should run automatically. You can verify by accessing Supabase Studio at:
- http://localhost:3000

Navigate to the SQL Editor and run:
```sql
SELECT * FROM entities;
```

You should see the default entities (server, ssl_certificate, database, application).

## Step 5: Create Admin User

1. In Supabase Studio (http://localhost:3000), go to Authentication → Users
2. Click "Add User" (or use the API to create a user)
3. Note the User ID
4. In SQL Editor, run:

```sql
INSERT INTO profiles (id, email, role, full_name)
VALUES ('YOUR_USER_ID_HERE', 'admin@example.com', 'admin', 'Admin User');
```

## Step 6: Start Development Server

```bash
npm run dev
```

Visit http://localhost:3001 (note: port 3001, as port 3000 is used by Supabase Studio)

## Step 7: Login

Use the email and password you created in Step 5 to log in.

## Troubleshooting

### Cannot connect to Supabase
- Ensure all Docker containers are running: `docker-compose ps`
- Check logs: `docker-compose logs`
- Verify environment variables are correctly set

### Migrations not applied
- Stop containers: `docker-compose down`
- Remove volumes: `docker volume rm infra-pm_db-data`
- Restart: `docker-compose up -d`

### Port conflicts
If port 3000, 5432, or 8000 are already in use:
- Edit `docker-compose.yaml` and change the port mappings
- Update `.env.local` with the new ports
- Restart: `docker-compose restart`

## Next Steps

Now that InfraPM is running:
1. Explore the default entities (Server, SSL Certificate, etc.)
2. Create your first record
3. Try creating a custom entity with your own fields
4. Invite other users (admin only)

For more details, see the main [README.md](README.md).
