# InfraPM - Dynamic Infrastructure Management System

A powerful, metadata-driven infrastructure management platform built with Next.js, TypeScript, and Supabase. InfraPM allows you to define custom entities and attributes dynamically, similar to a simplified CMDB or Airtable, specifically designed for managing infrastructure resources.

## Features

### 🎯 Core Features
- **Dynamic Schema Engine**: Define custom entities (Servers, SSL Certificates, etc.) and attributes on-the-fly
- **Flexible Field Types**: Support for String, Number, Date, and Relation field types
- **Granular RBAC**: Role-based access control with Admin, Maintainer, and Viewer roles
- **Entity-Level Permissions**: Fine-grained permissions per entity type
- **Audit Logging**: Comprehensive audit trail for all CRUD operations
- **Delete Protection**: Prevents deletion of records referenced by other records
- **Relation Validation**: Ensures data integrity for relational fields

### 🎨 UI/UX
- **Enterprise-Grade Design**: Clean, professional interface using shadcn/ui components
- **Dark/Light Mode**: Built-in theme switching
- **Advanced Tables**: TanStack Table with pagination (50/100/1000/All), filtering, and sorting
- **Responsive Design**: Works seamlessly across desktop and mobile devices

### 🔒 Security
- **Row Level Security (RLS)**: Database-level security policies
- **Supabase Auth**: Secure authentication system
- **Disabled Public Signup**: Only admins can create new users
- **Session Management**: Automatic session refresh and management

## Tech Stack

- **Frontend**: Next.js 16 (App Router), React 19, TypeScript
- **Styling**: Tailwind CSS v4, shadcn/ui
- **Backend**: Supabase (PostgreSQL)
- **State Management**: React Hooks
- **Form Handling**: React Hook Form + Zod validation
- **Tables**: TanStack Table
- **Testing**: Vitest + Testing Library
- **Containerization**: Docker Compose

## Prerequisites

- Node.js 20.x or later
- Docker and Docker Compose (for self-hosted Supabase)
- npm or pnpm

## Quick Start

### 1. Clone and Install Dependencies

```bash
cd infra-pm
npm install
```

### 2. Environment Setup

Copy the example environment file and configure it:

```bash
cp .env.example .env.local
```

Update the following variables in `.env.local`:
- `POSTGRES_PASSWORD`: Set a secure password
- `JWT_SECRET`: Generate a secure secret (min 32 characters)
- `SECRET_KEY_BASE`: Generate a secure key (min 64 characters)
- Configure SMTP settings for email functionality

### 3. Start Supabase Services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database (port 5432)
- Supabase Studio (port 3000)
- Kong API Gateway (port 8000)
- Auth server (GoTrue)
- PostgREST API
- Realtime server
- Storage server

### 4. Run Database Migrations

The migrations will run automatically when the database container starts. They are located in `supabase/migrations/`:
- `001_initial_schema.sql`: Creates all tables, RLS policies, and functions
- `002_seed_data.sql`: Seeds default entities and attributes

### 5. Create Admin User

Access Supabase Studio at `http://localhost:3000` and:
1. Navigate to Authentication → Users
2. Create a new user with admin email
3. Note the user ID
4. Execute in SQL Editor:

```sql
INSERT INTO profiles (id, email, role, full_name)
VALUES ('USER_ID_HERE', 'admin@example.com', 'admin', 'Admin User');
```

### 6. Start Development Server

```bash
npm run dev
```

Visit `http://localhost:3001` and log in with your admin credentials.

## Project Structure

```
infra-pm/
├── app/                    # Next.js App Router pages
│   ├── (auth)/            # Authentication pages
│   ├── (dashboard)/       # Main application pages
│   └── api/               # API routes
├── components/            # React components
│   ├── ui/               # shadcn/ui components
│   └── ...               # Feature components
├── lib/                   # Utility functions
│   ├── supabase/         # Supabase client utilities
│   ├── rbac/             # RBAC logic
│   └── validators/       # Validation utilities
├── types/                 # TypeScript type definitions
├── hooks/                 # Custom React hooks
├── supabase/             # Supabase configuration
│   ├── migrations/       # SQL migration files
│   └── kong.yml          # Kong API Gateway config
├── tests/                 # Test files
└── docker-compose.yaml    # Docker services configuration
```

## Database Schema

### Core Tables

#### RBAC Tables
- **profiles**: User profiles with role assignments
- **roles**: Role definitions (admin, maintainer, viewer)
- **permissions**: Entity-level permissions per role

#### Dynamic Schema Tables
- **entities**: Entity type definitions (e.g., "Server")
- **attributes**: Field definitions for entities
- **records**: Entity instances
- **values**: Attribute values for each record

#### Audit Trail
- **audit_logs**: Complete audit trail of all operations

### Entity Relationships

```
entities (1) ----< (N) attributes
                         |
entities (1) ----< (N) records ----< (N) values >---- (1) attributes
                         |
records (1) ----< (N) values (as relations)
```

## Testing

Run the test suite:

```bash
# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm run test:coverage
```

### Test Coverage

- RBAC permission logic
- Delete protection validation
- Record reference validation

## Development

### Adding New Entities (via UI)

1. Navigate to **Entities** in the dashboard
2. Click **New Entity**
3. Fill in entity details (name, display name, icon)
4. Add custom attributes with appropriate types
5. Save to create the entity schema

### Managing Records

1. Select an entity from the dashboard
2. View all records in a filterable, sortable table
3. Create, edit, or delete records (based on permissions)
4. The system enforces delete protection automatically

### Customizing Permissions

Admins can:
1. Navigate to **Permissions**
2. Select a role
3. Configure entity-level permissions
4. Save changes (applied immediately)

## Deployment

### Production Checklist

- [ ] Change all default passwords and secrets
- [ ] Configure SMTP for email notifications
- [ ] Set up SSL/TLS certificates
- [ ] Configure backup strategy for PostgreSQL
- [ ] Set up monitoring and logging
- [ ] Review and adjust RLS policies
- [ ] Test all authentication flows
- [ ] Perform security audit

### Environment Variables

Ensure these are properly set in production:
- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `SECRET_KEY_BASE`
- `NEXT_PUBLIC_SUPABASE_URL`
- `NEXT_PUBLIC_SUPABASE_ANON_KEY`
- SMTP configuration

## API Routes

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `POST /api/auth/refresh` - Refresh session

### Entities
- `GET /api/entities` - List all entities
- `POST /api/entities` - Create new entity (admin/maintainer)
- `GET /api/entities/:id` - Get entity details
- `PUT /api/entities/:id` - Update entity (admin/maintainer)
- `DELETE /api/entities/:id` - Delete entity (admin only)

### Records
- `GET /api/records/:entityId` - List records for entity
- `POST /api/records/:entityId` - Create record
- `GET /api/records/:entityId/:id` - Get record details
- `PUT /api/records/:entityId/:id` - Update record
- `DELETE /api/records/:entityId/:id` - Delete record (checks references)

### Audit Logs
- `GET /api/audit-logs` - List audit logs (filterable)

## Security Considerations

### Row Level Security (RLS)

All tables have RLS policies enabled:
- Users can only view data they have permission to access
- Admin role bypasses most restrictions
- Maintainers cannot delete records
- Viewers have read-only access

### Authentication

- Public signup is disabled by default
- Only admins can create new users
- Sessions are automatically refreshed
- JWT tokens are securely managed

### Data Validation

- All forms use Zod for validation
- Database constraints enforce data integrity
- Relation validation ensures referenced records exist
- Delete protection prevents orphaned references

## Troubleshooting

### Cannot connect to Supabase

1. Ensure Docker services are running: `docker-compose ps`
2. Check logs: `docker-compose logs`
3. Verify environment variables in `.env.local`
4. Ensure Kong Gateway is accessible on port 8000

### Migrations not applied

1. Stop containers: `docker-compose down`
2. Remove volumes: `docker volume rm infra-pm_db-data`
3. Restart: `docker-compose up -d`

### Permission denied errors

1. Check user role in `profiles` table
2. Verify RLS policies are enabled
3. Ensure user is authenticated
4. Check entity-specific permissions

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write/update tests
5. Run tests: `npm test`
6. Submit a pull request

## License

MIT License - See LICENSE file for details

## Support

For issues, questions, or contributions, please open an issue on GitHub.

---

Built with ❤️ using Next.js, TypeScript, and Supabase
