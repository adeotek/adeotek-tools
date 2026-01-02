# Infrastructure Management Portal

A modern, enterprise-ready web application for managing infrastructure data including servers, SSL certificates, applications, and services. Built with Next.js and Supabase.

## Features

### Core Functionality
- **Multi-user system** with role-based access control (Admin, Maintainer, Viewer)
- **Dynamic schema management** - Create custom data models and fields via UI
- **Pre-built data models**: Servers, SSL Certificates, Applications, Services
- **Audit logging** - Track all changes with user attribution
- **Relational integrity** - Prevent deletion of resources with dependencies
- **Self-hosted Supabase** - Complete backend infrastructure included

### Security
- Row Level Security (RLS) policies for all data
- Granular permissions per role and data model
- Secure authentication via Supabase Auth
- No public registration - admin controls all user creation
- Audit trail for compliance and tracking

### Technical Stack
- **Frontend**: Next.js 14 (App Router), React, TypeScript, Tailwind CSS
- **Backend**: Supabase (PostgreSQL, Auth, Real-time, Storage)
- **Database**: PostgreSQL 17 with RLS
- **Deployment**: Docker Compose
- **API**: Auto-generated REST and GraphQL APIs via PostgREST

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Node.js 20+ (for development)
- Git

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd adeotek-tools/infrastructure-management-portal
```

2. **Set up environment variables**
```bash
cp .env.example .env
```

3. **IMPORTANT: Update security credentials in `.env`**
```bash
# Generate secure values for these:
POSTGRES_PASSWORD=<your-secure-password>
JWT_SECRET=<your-jwt-secret-min-32-chars>
SECRET_KEY_BASE=<your-secret-key-base-min-64-chars>
```

4. **Start the application**
```bash
docker-compose up -d
```

This will start:
- PostgreSQL database (port 5432)
- Supabase services (Kong API Gateway on port 8000)
- Supabase Studio (port 3001) - Database management UI
- Next.js application (port 3002)

5. **Access the applications**
- **Main Application**: http://localhost:3002
- **Supabase Studio**: http://localhost:3001
- **API**: http://localhost:8000

6. **Default Admin Login**
```
Email: admin@example.com
Password: ChangeMe123!
```

**⚠️ IMPORTANT: Change the admin password immediately after first login!**

## Development

### Run locally without Docker

1. **Start Supabase services**
```bash
docker-compose up db kong auth rest realtime storage meta studio -d
```

2. **Install dependencies**
```bash
npm install
```

3. **Create local .env.local file**
```bash
NEXT_PUBLIC_SUPABASE_URL=http://localhost:8000
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0
```

4. **Run development server**
```bash
npm run dev
```

The application will be available at http://localhost:3000

### Project Structure
```
infrastructure-management-portal/
├── src/
│   ├── app/                    # Next.js app router pages
│   │   ├── auth/              # Authentication pages
│   │   ├── dashboard/         # Dashboard pages
│   │   └── api/               # API routes
│   ├── components/            # React components
│   ├── lib/                   # Utilities and helpers
│   │   ├── supabase/         # Supabase client configuration
│   │   └── types/            # TypeScript type definitions
│   └── proxy.ts              # Next.js proxy for auth
├── supabase/
│   ├── config/               # Supabase configuration
│   │   └── kong.yml         # API Gateway configuration
│   └── migrations/          # Database migrations
│       ├── 00001_initial_schema.sql
│       └── 00002_seed_admin_user.sql
├── docker-compose.yml        # Docker services configuration
├── Dockerfile               # Next.js application container
└── .env.example            # Environment variables template
```

## Database Schema

### Core Tables

#### Users and Roles
- `imp.roles` - Role definitions (Admin, Maintainer, Viewer)
- `imp.user_profiles` - User profiles extending Supabase auth
- `imp.permissions` - Role-based permissions per data model

#### Dynamic Schema
- `imp.data_models` - Metadata for data models
- `imp.field_types` - Available field types
- `imp.fields` - Field definitions for each model

#### Infrastructure Data
- `imp.servers` - Server inventory
- `imp.ssl_certificates` - SSL certificate tracking
- `imp.applications` - Application catalog
- `imp.services` - Services running on servers

#### Audit
- `imp.audit_logs` - Complete audit trail

### Role Permissions

| Role | Read | Create | Update | Delete |
|------|------|--------|--------|--------|
| Admin | ✅ | ✅ | ✅ | ✅ |
| Maintainer | ✅ | ✅ | ✅ | ❌ |
| Viewer | ✅ | ❌ | ❌ | ❌ |

Permissions are configurable per role and per data model.

## API Access

The application provides REST and GraphQL APIs via PostgREST:

### REST API
```bash
# Get all servers
curl http://localhost:8000/rest/v1/servers \
  -H "apikey: YOUR_ANON_KEY" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### GraphQL API
Available at: http://localhost:8000/graphql/v1

## Production Deployment

### Security Checklist

- [ ] Change all default passwords
- [ ] Generate new JWT tokens with strong secret
- [ ] Update POSTGRES_PASSWORD
- [ ] Update SECRET_KEY_BASE
- [ ] Configure SMTP for email notifications
- [ ] Enable SSL/TLS for all services
- [ ] Set up firewall rules
- [ ] Enable database backups
- [ ] Configure log rotation
- [ ] Review and update CORS settings in kong.yml

### Backup Strategy

1. **Database Backups**
```bash
# Backup database
docker-compose exec db pg_dump -U postgres postgres > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore database
docker-compose exec -T db psql -U postgres postgres < backup_file.sql
```

## Troubleshooting

### Database connection issues
```bash
# Check database logs
docker-compose logs db

# Verify database is healthy
docker-compose ps
```

### Authentication issues
```bash
# Check auth service logs
docker-compose logs auth

# Verify JWT_SECRET is consistent across services
```

## License

See repository LICENSE file.
