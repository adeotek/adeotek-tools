# InfraPM Project Structure

This document describes the organization and architecture of the InfraPM project.

## Directory Structure

```
infra-pm/
├── app/                          # Next.js App Router
│   ├── (auth)/                   # Authentication routes group
│   │   ├── login/               # Login page
│   │   └── layout.tsx           # Auth layout (no sidebar)
│   ├── (dashboard)/             # Main application routes group
│   │   ├── entities/            # Entity management pages
│   │   ├── records/             # Record management pages
│   │   ├── audit-logs/          # Audit log viewer
│   │   ├── users/               # User management (admin only)
│   │   └── layout.tsx           # Dashboard layout (with sidebar)
│   ├── api/                     # API routes
│   │   ├── entities/            # Entity CRUD endpoints
│   │   ├── records/             # Record CRUD endpoints
│   │   ├── audit-logs/          # Audit log endpoints
│   │   └── auth/                # Auth endpoints
│   ├── globals.css              # Global styles with CSS variables
│   ├── layout.tsx               # Root layout
│   └── page.tsx                 # Home page
│
├── components/                   # React components
│   ├── ui/                      # shadcn/ui components
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── dialog.tsx
│   │   ├── form.tsx
│   │   ├── input.tsx
│   │   ├── table.tsx
│   │   └── ...
│   ├── entities/                # Entity-specific components
│   │   ├── entity-form.tsx
│   │   ├── entity-list.tsx
│   │   └── attribute-editor.tsx
│   ├── records/                 # Record-specific components
│   │   ├── record-form.tsx
│   │   ├── record-table.tsx
│   │   └── value-editor.tsx
│   ├── layout/                  # Layout components
│   │   ├── sidebar.tsx
│   │   ├── header.tsx
│   │   └── theme-toggle.tsx
│   └── shared/                  # Shared components
│       ├── loading.tsx
│       ├── error-boundary.tsx
│       └── pagination.tsx
│
├── lib/                         # Utility functions and logic
│   ├── supabase/               # Supabase client utilities
│   │   ├── client.ts           # Browser client
│   │   ├── server.ts           # Server client
│   │   └── middleware.ts       # Middleware helper
│   ├── rbac/                   # Role-based access control
│   │   └── permissions.ts      # Permission checking logic
│   ├── validators/             # Validation utilities
│   │   └── record-validator.ts # Record and relation validation
│   └── utils.ts                # General utilities (cn, etc.)
│
├── types/                       # TypeScript type definitions
│   ├── database.ts             # Supabase database types
│   └── index.ts                # Exported types
│
├── hooks/                       # Custom React hooks
│   ├── use-entities.ts         # Entity management hooks
│   ├── use-records.ts          # Record management hooks
│   ├── use-auth.ts             # Authentication hooks
│   └── use-permissions.ts      # Permission checking hooks
│
├── supabase/                    # Supabase configuration
│   ├── migrations/             # SQL migration files
│   │   ├── 001_initial_schema.sql
│   │   └── 002_seed_data.sql
│   └── kong.yml                # Kong API Gateway config
│
├── tests/                       # Test files
│   ├── setup.ts                # Test setup
│   ├── rbac.test.ts            # RBAC tests
│   └── validators.test.ts      # Validator tests
│
├── public/                      # Static assets
│   └── ...
│
├── docker-compose.yaml          # Supabase services
├── .env.example                 # Environment variables template
├── .env.local                   # Local environment (not committed)
├── components.json              # shadcn/ui config
├── middleware.ts                # Next.js middleware
├── next.config.ts               # Next.js configuration
├── package.json                 # Dependencies and scripts
├── tsconfig.json                # TypeScript configuration
├── vitest.config.ts             # Vitest configuration
├── README.md                    # Main documentation
└── SETUP.md                     # Quick setup guide
```

## Key Architectural Patterns

### 1. Metadata-Driven Schema

The core of InfraPM is its metadata-driven architecture:

- **entities**: Defines types (e.g., "Server", "SSL Certificate")
- **attributes**: Defines fields for each entity type
- **records**: Stores instances of entities
- **values**: Stores the actual data for each field

This allows users to define custom schemas without code changes.

### 2. Role-Based Access Control (RBAC)

Three levels of access:
- **Admin**: Full CRUD access
- **Maintainer**: Create, Read, Update (no Delete)
- **Viewer**: Read-only

Permissions can be:
- Global (applies to all entities)
- Entity-specific (per entity type)

### 3. Row Level Security (RLS)

All security is enforced at the database level using Supabase RLS policies:
- Policies defined in migration files
- Cannot be bypassed by application code
- Automatically applied to all queries

### 4. Audit Logging

All CRUD operations are automatically logged:
- Database triggers capture changes
- Stored in `audit_logs` table
- Includes old/new values for updates

### 5. Delete Protection

Records cannot be deleted if referenced by other records:
- Checked via `check_record_references()` function
- Prevents orphaned references
- Returns list of blocking references

## Database Schema

### RBAC Tables

```
profiles (extends auth.users)
  ├── id (FK to auth.users)
  ├── email
  ├── role (admin/maintainer/viewer)
  └── timestamps

roles
  ├── id
  ├── name
  └── description

permissions
  ├── id
  ├── role_id (FK to roles)
  ├── entity_type_id (FK to entities, NULL = all)
  ├── can_create, can_read, can_update, can_delete
  └── timestamp
```

### Dynamic Schema Tables

```
entities
  ├── id
  ├── name (unique)
  ├── display_name
  ├── description
  ├── icon
  └── timestamps

attributes
  ├── id
  ├── entity_id (FK to entities)
  ├── name
  ├── attribute_type (string/number/date/relation)
  ├── is_required
  ├── references_entity_id (FK to entities, for relations)
  ├── validation_rules (JSONB)
  ├── order_index
  └── timestamps

records
  ├── id
  ├── entity_id (FK to entities)
  ├── created_by (FK to auth.users)
  └── timestamps

values
  ├── id
  ├── record_id (FK to records)
  ├── attribute_id (FK to attributes)
  ├── value_text
  ├── value_number
  ├── value_date
  ├── value_relation (FK to records)
  └── timestamps
```

### Audit Table

```
audit_logs
  ├── id
  ├── user_id (FK to auth.users)
  ├── action (create/update/delete/view)
  ├── entity_type
  ├── record_id
  ├── old_values (JSONB)
  ├── new_values (JSONB)
  ├── ip_address
  ├── user_agent
  └── created_at
```

## Technology Stack

### Frontend
- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS v4
- **UI Components**: shadcn/ui
- **Tables**: TanStack Table
- **Forms**: React Hook Form + Zod
- **Theme**: next-themes (dark/light mode)

### Backend
- **Database**: PostgreSQL (via Supabase)
- **Auth**: Supabase Auth (GoTrue)
- **API**: Supabase PostgREST
- **Realtime**: Supabase Realtime
- **Storage**: Supabase Storage

### Development
- **Testing**: Vitest + Testing Library
- **Linting**: ESLint
- **Type Checking**: TypeScript
- **Package Manager**: npm
- **Container**: Docker Compose

## API Conventions

### RESTful Endpoints

```
GET    /api/entities          # List entities
POST   /api/entities          # Create entity
GET    /api/entities/:id      # Get entity
PUT    /api/entities/:id      # Update entity
DELETE /api/entities/:id      # Delete entity

GET    /api/records/:entityId          # List records
POST   /api/records/:entityId          # Create record
GET    /api/records/:entityId/:id      # Get record
PUT    /api/records/:entityId/:id      # Update record
DELETE /api/records/:entityId/:id      # Delete record

GET    /api/audit-logs        # List audit logs
```

### Response Format

Success:
```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "perPage": 50,
    "total": 100
  }
}
```

Error:
```json
{
  "error": {
    "message": "Error message",
    "code": "ERROR_CODE",
    "details": { ... }
  }
}
```

## Development Workflow

1. **Start Supabase**: `docker-compose up -d`
2. **Install dependencies**: `npm install`
3. **Run development**: `npm run dev`
4. **Run tests**: `npm test`
5. **Build**: `npm run build`
6. **Lint**: `npm run lint`

## Security Best Practices

1. **Never commit secrets**: Use `.env.local` for sensitive data
2. **RLS policies**: Always enable RLS on new tables
3. **Input validation**: Use Zod schemas for all user input
4. **Audit logging**: Log all sensitive operations
5. **Type safety**: Leverage TypeScript for compile-time checks
6. **Permission checks**: Always check permissions before operations

## Contributing

When adding new features:
1. Add types to `types/database.ts`
2. Add validation logic to `lib/validators/`
3. Add permission checks using `lib/rbac/permissions.ts`
4. Write tests in `tests/`
5. Update documentation

## Future Enhancements

Potential areas for expansion:
- [ ] Export/Import functionality
- [ ] API webhooks
- [ ] Custom workflows
- [ ] Advanced reporting
- [ ] Integration with monitoring tools
- [ ] Multi-tenancy support
- [ ] Advanced search/filtering
- [ ] Bulk operations
- [ ] File attachments
- [ ] Comments/notes on records
