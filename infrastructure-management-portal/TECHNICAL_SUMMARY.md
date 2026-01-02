# Infrastructure Management Portal - Technical Summary

## Project Overview
Enterprise-ready web application for managing infrastructure data with dynamic schema capabilities, built on Next.js 14 and self-hosted Supabase.

## Technology Stack

### Frontend
- **Framework**: Next.js 14.3 (App Router)
- **Language**: TypeScript 5.7
- **UI Library**: React 19
- **Styling**: Tailwind CSS 3.4
- **Forms**: React Hook Form + Zod validation
- **Icons**: Lucide React
- **Date Handling**: date-fns

### Backend
- **Database**: PostgreSQL 17.6
- **Auth**: Supabase GoTrue v2.162
- **API**: PostgREST v12.2 (Auto-generated REST API)
- **Real-time**: Supabase Realtime v2.30
- **Storage**: Supabase Storage v1.11
- **API Gateway**: Kong 2.8

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Image Processing**: imgproxy v3.25
- **Database UI**: Supabase Studio

## Architecture

### Application Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                         Client Browser                       │
│                   (Next.js React Application)                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP/WebSocket
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Kong API Gateway                        │
│                  (Port 8000 - API Routing)                  │
└──┬──────────┬──────────┬───────────┬───────────┬───────────┘
   │          │          │           │           │
   │ Auth     │ REST     │ Realtime  │ Storage   │ Meta
   ▼          ▼          ▼           ▼           ▼
┌─────┐  ┌──────┐  ┌──────────┐ ┌─────────┐ ┌────────┐
│GoTrue│ │PostgREST│ │Realtime│ │Storage│ │PG Meta│
└──┬──┘ └───┬────┘ └────┬─────┘ └────┬────┘ └───┬────┘
   │        │           │            │          │
   └────────┴───────────┴────────────┴──────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   PostgreSQL 17.6    │
              │   (Port 5432)        │
              │   - auth schema      │
              │   - public schema    │
              │   - imp schema       │
              │   - storage schema   │
              └──────────────────────┘
```

### Database Schema Architecture
```
imp schema (Infrastructure Management Portal)
├── Core Tables
│   ├── roles (system & custom roles)
│   ├── user_profiles (extends auth.users)
│   ├── permissions (role × model permissions)
│   └── audit_logs (complete audit trail)
│
├── Dynamic Schema System
│   ├── data_models (model definitions)
│   ├── field_types (available field types)
│   └── fields (model field definitions)
│
└── Core Data Models
    ├── servers (infrastructure servers)
    ├── ssl_certificates (SSL/TLS certificates)
    ├── applications (software applications)
    └── services (services on servers)
```

## Security Architecture

### Authentication Flow
```
1. User → Login Page (email + password)
2. Next.js → Supabase Auth (GoTrue)
3. GoTrue → PostgreSQL auth.users verification
4. GoTrue → Generate JWT with role claims
5. JWT → Stored in HTTP-only cookie
6. Subsequent requests → JWT validated by Kong → PostgREST
```

### Row Level Security (RLS)
Every table has RLS policies enforcing permissions:
- `SELECT`: Checks `has_permission(table_name, 'read')`
- `INSERT`: Checks `has_permission(table_name, 'create')`
- `UPDATE`: Checks `has_permission(table_name, 'update')`
- `DELETE`: Checks `has_permission(table_name, 'delete')`

### Permission Model
```sql
-- Helper function checks user's role permissions
has_permission(model_name TEXT, permission_type TEXT) → BOOLEAN

-- Permissions stored in imp.permissions table
-- Links: role_id → model_id → {can_read, can_create, can_update, can_delete}
```

### Audit Logging
Automatic audit trail via triggers:
- Captures: user_id, user_email, action, entity_type, entity_id
- Stores: old_values (JSON), new_values (JSON)
- Includes: ip_address, user_agent, timestamp
- Trigger: `imp.log_audit_trail()` on INSERT/UPDATE/DELETE

## Data Models

### Default Roles
| Role | Permissions | Notes |
|------|------------|-------|
| Admin | Full CRUD + Delete | Can manage users, models, permissions |
| Maintainer | Create + Read + Update | Cannot delete resources |
| Viewer | Read-only | Cannot modify any data |

### Core Data Models

#### Servers
```typescript
{
  hostname: string
  ip_address: string (INET)
  ram_gb: number
  cpu_cores: number
  cpu_model: string
  description: string
  group_name: string
  location: string
  os_name: string
  os_version: string
  status: string
  notes: string
  metadata: JSONB
}
```

#### SSL Certificates
```typescript
{
  name: string
  common_name: string
  valid_from: Date
  valid_to: Date
  sans: string[]
  issuer: string
  serial_number: string
  certificate_type: string
  status: string
  notes: string
  metadata: JSONB
}
```

#### Applications
```typescript
{
  name: string
  type: string
  group_name: string
  version: string
  description: string
  repository_url: string
  documentation_url: string
  status: string
  notes: string
  metadata: JSONB
}
```

#### Services
```typescript
{
  name: string
  type: string
  group_name: string
  server_id: UUID (FK → servers)
  application_id: UUID (FK → applications)
  fqdn: string
  ip_address: string (INET)
  port: number
  ssl_certificate_id: UUID (FK → ssl_certificates)
  protocol: string
  status: string
  health_check_url: string
  description: string
  notes: string
  metadata: JSONB
}
```

## API Endpoints

### REST API (PostgREST)
Base URL: `http://localhost:8000/rest/v1/`

All tables exposed as RESTful endpoints:
- `GET /servers` - List servers (with filtering, pagination)
- `POST /servers` - Create server
- `GET /servers?id=eq.{uuid}` - Get specific server
- `PATCH /servers?id=eq.{uuid}` - Update server
- `DELETE /servers?id=eq.{uuid}` - Delete server

### GraphQL API
Base URL: `http://localhost:8000/graphql/v1/`

Auto-generated GraphQL schema from database.

### Custom API Routes (Next.js)
- `GET /api/health` - Health check endpoint
- `GET /auth/callback` - OAuth callback handler

## Environment Configuration

### Required Environment Variables
```bash
# Database
POSTGRES_PASSWORD=<secure-password>

# JWT & Secrets
JWT_SECRET=<min-32-chars>
SECRET_KEY_BASE=<min-64-chars>
ANON_KEY=<jwt-token-for-anon-role>
SERVICE_ROLE_KEY=<jwt-token-for-service-role>

# URLs
API_EXTERNAL_URL=http://localhost:8000
SUPABASE_PUBLIC_URL=http://localhost:8000
SITE_URL=http://localhost:3002
NEXT_PUBLIC_SUPABASE_URL=http://localhost:8000
NEXT_PUBLIC_SUPABASE_ANON_KEY=<anon-key>

# Auth
DISABLE_SIGNUP=true
ENABLE_EMAIL_SIGNUP=true
ENABLE_EMAIL_AUTOCONFIRM=false

# SMTP (optional but recommended)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASS=<smtp-password>
```

## Deployment

### Docker Compose Services
| Service | Port | Purpose |
|---------|------|---------|
| db | 5432 | PostgreSQL database |
| kong | 8000 | API Gateway |
| auth | 9999 | Authentication service |
| rest | 3000 | REST API (PostgREST) |
| realtime | 4000 | Real-time subscriptions |
| storage | 5000 | File storage API |
| studio | 3001 | Database management UI |
| meta | 8080 | Database metadata API |
| imgproxy | 5001 | Image transformation |
| app | 3002 | Next.js application |

### Volumes
- `db-data`: PostgreSQL data persistence
- `storage-data`: Uploaded files storage

### Networks
- `imp-network`: Bridge network for all services

## Development Workflow

### Local Development
```bash
# 1. Start Supabase services only
docker-compose up db kong auth rest realtime storage meta studio -d

# 2. Install dependencies
npm install

# 3. Create .env.local
cp .env.local.example .env.local

# 4. Run Next.js dev server
npm run dev

# Access at http://localhost:3000
```

### Production Build
```bash
# Full stack with Docker
docker-compose up -d

# Access at http://localhost:3002
```

## Testing Strategy

### Unit Tests (To Be Implemented)
- Database functions: `has_permission()`, `get_user_role()`
- Schema migration utilities
- Form validation logic
- API route handlers

### Integration Tests (To Be Implemented)
- Full CRUD workflows
- Permission enforcement
- Audit log creation
- Relationship integrity
- Authentication flow

### E2E Tests (To Be Implemented)
- User login/logout
- Create/edit/delete resources
- Role-based access scenarios
- Dynamic schema creation

## Performance Considerations

### Database
- Indexes on frequently queried columns (hostname, IP, FQDN)
- Connection pooling via PgBouncer (can be added)
- Query optimization with EXPLAIN ANALYZE

### Application
- Next.js static generation where possible
- Server-side rendering for dynamic content
- Incremental Static Regeneration (ISR)
- Image optimization via imgproxy

### Caching
- Browser cache for static assets
- Redis can be added for session storage
- PostgREST response caching

## Monitoring & Observability

### Health Checks
- Application: `GET /api/health`
- Database: PostgreSQL health check
- Services: Docker health check directives

### Logging
- Application logs: Next.js stdout/stderr
- Database logs: PostgreSQL logs
- Audit logs: `imp.audit_logs` table
- Kong access logs: API request logs

### Metrics (To Be Added)
- Request rate and latency
- Database query performance
- Error rates
- User activity metrics

## Security Best Practices

### Implemented
✅ Row Level Security on all tables
✅ JWT-based authentication
✅ Password hashing (bcrypt via Supabase)
✅ CORS configuration
✅ SQL injection prevention (parameterized queries)
✅ XSS prevention (React escaping)
✅ Audit logging

### Recommended for Production
⚠️ Enable HTTPS/TLS with valid certificates
⚠️ Implement rate limiting
⚠️ Set up WAF (Web Application Firewall)
⚠️ Enable database encryption at rest
⚠️ Implement secrets management (Vault, etc.)
⚠️ Set up intrusion detection
⚠️ Regular security audits
⚠️ Dependency vulnerability scanning

## Backup & Recovery

### Database Backups
```bash
# Automated backup script
docker-compose exec db pg_dump -U postgres postgres | gzip > backup_$(date +%Y%m%d).sql.gz

# Restore
gunzip < backup_20260102.sql.gz | docker-compose exec -T db psql -U postgres postgres
```

### Storage Backups
```bash
# Backup storage volume
docker run --rm -v imp_storage-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/storage_$(date +%Y%m%d).tar.gz /data
```

## Future Enhancements

### Phase 1 (UI Completion)
- [ ] Dynamic schema builder UI
- [ ] CRUD interfaces for core models
- [ ] User management UI
- [ ] Permission configuration UI
- [ ] Audit log viewer

### Phase 2 (Advanced Features)
- [ ] Import/export (CSV, JSON)
- [ ] Advanced filtering & search
- [ ] Custom dashboards
- [ ] Reporting engine
- [ ] API key management
- [ ] Webhooks

### Phase 3 (Enterprise)
- [ ] SSO/SAML integration
- [ ] Multi-tenancy
- [ ] Field-level encryption
- [ ] Advanced compliance reports
- [ ] Data retention policies
- [ ] Workflow automation

## Troubleshooting

### Common Issues

**Database Connection Failed**
```bash
# Check database is running
docker-compose ps db

# Check database logs
docker-compose logs db

# Verify credentials in .env
```

**Authentication Errors**
```bash
# Verify JWT_SECRET matches across services
# Check auth service logs
docker-compose logs auth

# Verify ANON_KEY and SERVICE_ROLE_KEY are correct
```

**Permission Denied Errors**
```bash
# Check RLS policies
# Verify user has correct role
# Check permissions table for role-model combination
```

## Contributing

See `IMPLEMENTATION_GUIDE.md` for detailed development instructions.

## License

See repository LICENSE file.

## Support & Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [Supabase Documentation](https://supabase.com/docs)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [PostgREST Documentation](https://postgrest.org/)
- Project README: `README.md`
- Implementation Guide: `IMPLEMENTATION_GUIDE.md`
