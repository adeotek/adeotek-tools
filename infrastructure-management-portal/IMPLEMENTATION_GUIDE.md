# Implementation Guide for Dynamic Schema Builder

This document outlines the implementation plan for the remaining features of the Infrastructure Management Portal, particularly the UI-based dynamic schema builder.

## Current State

### ✅ Completed
- Database schema with dynamic tables (`data_models`, `fields`, `field_types`)
- Row Level Security (RLS) policies
- Core data models (Servers, SSL Certs, Applications, Services)
- Authentication system
- Basic dashboard
- Docker Compose configuration
- API routes foundation

### 🚧 Remaining Work

## Phase 1: Dynamic Schema Builder UI

### 1.1 Data Models Management Page
**Location**: `src/app/dashboard/data-models/page.tsx`

**Features**:
- List all data models with their metadata
- Filter by system/custom models
- Search by name
- Display model statistics (record count, field count)

**Components needed**:
- `DataModelsList` - Grid or table of models
- `DataModelCard` - Individual model display
- `CreateModelButton` - Opens create modal

### 1.2 Create/Edit Data Model Modal
**Component**: `src/components/data-models/DataModelForm.tsx`

**Fields**:
- Model name (technical identifier)
- Display name (user-friendly)
- Description
- Icon selection
- Active/Inactive toggle

**Validation**:
- Name must be unique and slug-friendly
- Table name auto-generated from name
- Display name required

### 1.3 Field Management Interface
**Component**: `src/components/data-models/FieldManagement.tsx`

**Features**:
- Add/edit/delete fields for a model
- Drag-and-drop field ordering
- Field configuration panel

**Field Configuration Options**:
- Field name & display name
- Field type (from `field_types` table)
- Required/Optional
- Unique constraint
- Default value
- Reference model (for relationships)
- Validation rules (JSON)

### 1.4 Schema Migration Engine
**Location**: `src/lib/schema-migration.ts`

**Functions needed**:
```typescript
// Generate SQL for creating a new table
async function createTableFromModel(model: DataModel, fields: Field[]): Promise<string>

// Generate SQL for adding a field
async function addFieldToTable(tableName: string, field: Field): Promise<string>

// Generate SQL for modifying a field
async function modifyField(tableName: string, field: Field, oldField: Field): Promise<string>

// Generate SQL for deleting a field
async function deleteField(tableName: string, fieldName: string): Promise<string>

// Execute migration with transaction
async function executeMigration(sql: string): Promise<void>
```

**Important**: All custom tables should be created in the `imp` schema and have RLS enabled.

## Phase 2: CRUD Interfaces for Core Models

### 2.1 Generic CRUD Component
**Component**: `src/components/crud/GenericCRUD.tsx`

This component should dynamically render CRUD interfaces based on model schema.

**Props**:
```typescript
interface GenericCRUDProps {
  modelId: string
  modelName: string
  fields: Field[]
  permissions: Permission
}
```

**Features**:
- Data grid with sorting, filtering, pagination
- Add new record button
- Edit record modal
- Delete with confirmation
- Export to CSV/JSON
- Search across all fields

### 2.2 Dynamic Form Generator
**Component**: `src/components/forms/DynamicForm.tsx`

**Input Props**:
```typescript
interface DynamicFormProps {
  fields: Field[]
  initialData?: Record<string, any>
  onSubmit: (data: Record<string, any>) => Promise<void>
  mode: 'create' | 'edit'
}
```

**Field Type Renderers**:
- `text` → `<input type="text">`
- `text_long` → `<textarea>`
- `number` → `<input type="number">`
- `boolean` → `<input type="checkbox">`
- `date` → `<input type="date">`
- `datetime` → Date/time picker component
- `email` → `<input type="email">`
- `url` → `<input type="url">`
- `reference` → Dropdown/autocomplete from referenced table
- `json` → JSON editor component

### 2.3 Pages for Core Models
Create these pages:
- `src/app/dashboard/servers/page.tsx`
- `src/app/dashboard/ssl-certificates/page.tsx`
- `src/app/dashboard/applications/page.tsx`
- `src/app/dashboard/services/page.tsx`

Each should use `GenericCRUD` component with model-specific configuration.

## Phase 3: Admin Features

### 3.1 User Management
**Location**: `src/app/dashboard/users/page.tsx`

**Features**:
- List all users with roles
- Create new user (sends invitation email)
- Edit user profile and role
- Deactivate/reactivate users
- Password reset option

**API Routes needed**:
- `POST /api/admin/users` - Create user
- `PUT /api/admin/users/[id]` - Update user
- `POST /api/admin/users/[id]/reset-password` - Send reset email

### 3.2 Permission Configuration
**Location**: `src/app/dashboard/permissions/page.tsx`

**Features**:
- Matrix view: Roles (columns) × Models (rows)
- Checkboxes for each permission (Read/Create/Update/Delete)
- Bulk permission updates
- Preview mode to see effective permissions

### 3.3 Audit Log Viewer
**Location**: `src/app/dashboard/audit-logs/page.tsx`

**Features**:
- Filterable log table
- Filters: user, date range, action type, entity type
- View old/new values diff
- Export logs
- Retention policy configuration

## Phase 4: Data Model Builder UI

### 4.1 Model Designer Interface
**Component**: `src/components/schema-builder/ModelDesigner.tsx`

Visual interface for building models:
1. Model properties panel (left)
2. Fields list (center) with drag-drop
3. Field properties panel (right)

**Actions**:
- Add field button → Opens field selector modal
- Each field shows: type icon, name, required indicator, actions
- Click field → Opens properties panel
- Delete field → Confirmation dialog
- Save model → Validates and creates table

### 4.2 Field Type Selector
**Component**: `src/components/schema-builder/FieldTypeSelector.tsx`

**Display**:
- Grid of available field types with icons
- Search/filter by name
- Category grouping (Text, Numbers, Dates, Relations)

### 4.3 Relationship Builder
**Component**: `src/components/schema-builder/RelationshipBuilder.tsx`

For reference fields, configure:
- Target model selection
- Display field (which field to show in dropdown)
- On delete behavior (CASCADE, SET NULL, RESTRICT)

## Phase 5: Testing & Polish

### 5.1 Unit Tests
Test coverage needed for:
- Schema migration functions
- Permission checks
- Audit log triggers
- CRUD operations

### 5.2 Integration Tests
- Full workflow: Create model → Add fields → Create record → Edit → Delete
- Permission enforcement tests
- Relationship integrity tests

### 5.3 UI Polish
- Loading states
- Error handling & user-friendly messages
- Responsive design for mobile
- Accessibility (ARIA labels, keyboard navigation)
- Dark mode support (optional)

## Technical Considerations

### Database Schema Changes
When creating new tables dynamically:

```sql
-- Template for new table
CREATE TABLE imp.{table_name} (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- Dynamic fields here --
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id),
    updated_by UUID REFERENCES auth.users(id)
);

-- Enable RLS
ALTER TABLE imp.{table_name} ENABLE ROW LEVEL SECURITY;

-- Create policies (read/create/update/delete)
CREATE POLICY "Users can read if permitted" ON imp.{table_name}
    FOR SELECT USING (imp.has_permission('imp.{table_name}', 'read'));
-- ... similar for other operations

-- Create audit trigger
CREATE TRIGGER audit_{table_name} 
    AFTER INSERT OR UPDATE OR DELETE ON imp.{table_name}
    FOR EACH ROW EXECUTE FUNCTION imp.log_audit_trail();

-- Create updated_at trigger
CREATE TRIGGER update_{table_name}_updated_at 
    BEFORE UPDATE ON imp.{table_name}
    FOR EACH ROW EXECUTE FUNCTION imp.update_updated_at_column();
```

### Security Considerations
1. Always validate field names to prevent SQL injection
2. Use parameterized queries or Supabase client for data operations
3. Limit schema modifications to admin role only
4. Log all schema changes in audit log
5. Backup database before applying migrations

### Performance Optimization
1. Add indexes on frequently queried fields
2. Implement pagination for large datasets
3. Use Supabase real-time subscriptions for live updates
4. Cache field type definitions
5. Lazy load models and fields in dropdowns

## Deployment Checklist

Before production deployment:
- [ ] Change all default passwords
- [ ] Generate production JWT secret
- [ ] Configure SMTP for emails
- [ ] Set up SSL/TLS certificates
- [ ] Enable database backups
- [ ] Configure monitoring and alerts
- [ ] Set up log aggregation
- [ ] Test disaster recovery procedures
- [ ] Document admin procedures
- [ ] Create user training materials

## API Documentation

Auto-generate API docs using:
- OpenAPI spec from PostgREST
- Add Swagger UI at `/api/docs`
- Document custom endpoints

## Future Enhancements

### Phase 6: Advanced Features
- Import/export data (CSV, JSON, Excel)
- Bulk operations
- Advanced filters and saved views
- Custom reports and dashboards
- API keys for external integrations
- Webhooks for data changes
- Data validation rules engine
- Workflow automation
- Multi-tenancy support
- Activity timeline for records

### Phase 7: Enterprise Features
- Single Sign-On (SSO) integration
- LDAP/Active Directory support
- Advanced audit compliance reports
- Data encryption at rest
- Field-level encryption for sensitive data
- Granular field-level permissions
- Data retention policies
- Compliance reporting (SOC 2, ISO 27001)

## Resources

- [Supabase Documentation](https://supabase.com/docs)
- [Next.js App Router](https://nextjs.org/docs/app)
- [PostgreSQL Dynamic SQL](https://www.postgresql.org/docs/current/plpgsql-statements.html#PLPGSQL-STATEMENTS-EXECUTING-DYN)
- [Row Level Security Guide](https://supabase.com/docs/guides/auth/row-level-security)

## Support

For implementation questions or issues:
1. Check Supabase Discord community
2. Review PostgreSQL documentation for RLS
3. Next.js GitHub discussions for frontend issues
