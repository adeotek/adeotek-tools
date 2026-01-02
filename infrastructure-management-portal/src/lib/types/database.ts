export interface Role {
  id: string
  name: string
  description: string | null
  is_system_role: boolean
  created_at: string
  updated_at: string
}

export interface UserProfile {
  id: string
  role_id: string
  full_name: string | null
  email: string
  is_active: boolean
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
  role?: Role
}

export interface DataModel {
  id: string
  name: string
  display_name: string
  description: string | null
  table_name: string
  icon: string | null
  is_system_model: boolean
  is_active: boolean
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

export interface FieldType {
  id: string
  name: string
  display_name: string
  sql_type: string
  validation_rules: any
  ui_component: string | null
  description: string | null
}

export interface Field {
  id: string
  model_id: string
  name: string
  display_name: string
  field_type_id: string
  is_required: boolean
  is_unique: boolean
  default_value: string | null
  validation_rules: any
  reference_model_id: string | null
  display_order: number
  description: string | null
  is_system_field: boolean
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
  field_type?: FieldType
  reference_model?: DataModel
}

export interface Permission {
  id: string
  role_id: string
  model_id: string
  can_read: boolean
  can_create: boolean
  can_update: boolean
  can_delete: boolean
  created_at: string
  updated_at: string
}

export interface AuditLog {
  id: string
  user_id: string | null
  user_email: string | null
  action: string
  entity_type: string
  entity_id: string | null
  old_values: any
  new_values: any
  ip_address: string | null
  user_agent: string | null
  created_at: string
}

// Core data models
export interface Server {
  id: string
  hostname: string
  ip_address: string
  ram_gb: number | null
  cpu_cores: number | null
  cpu_model: string | null
  description: string | null
  group_name: string | null
  location: string | null
  os_name: string | null
  os_version: string | null
  status: string
  notes: string | null
  metadata: any
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

export interface SSLCertificate {
  id: string
  name: string
  common_name: string
  valid_from: string
  valid_to: string
  sans: string[] | null
  issuer: string | null
  serial_number: string | null
  certificate_type: string | null
  status: string
  notes: string | null
  metadata: any
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

export interface Application {
  id: string
  name: string
  type: string | null
  group_name: string | null
  version: string | null
  description: string | null
  repository_url: string | null
  documentation_url: string | null
  status: string
  notes: string | null
  metadata: any
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

export interface Service {
  id: string
  name: string
  type: string | null
  group_name: string | null
  server_id: string | null
  application_id: string | null
  fqdn: string | null
  ip_address: string | null
  port: number | null
  ssl_certificate_id: string | null
  protocol: string | null
  status: string
  health_check_url: string | null
  description: string | null
  notes: string | null
  metadata: any
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
  server?: Server
  application?: Application
  ssl_certificate?: SSLCertificate
}

// Database types for RPC calls
export interface Database {
  public: {
    Tables: {}
    Views: {}
    Functions: {
      has_permission: {
        Args: { model_name: string; permission_type: string }
        Returns: boolean
      }
      get_user_role: {
        Args: {}
        Returns: string
      }
    }
    Enums: {}
  }
  imp: {
    Tables: {
      roles: {
        Row: Role
        Insert: Omit<Role, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<Role, 'id' | 'created_at' | 'updated_at'>>
      }
      user_profiles: {
        Row: UserProfile
        Insert: Omit<UserProfile, 'created_at' | 'updated_at'>
        Update: Partial<Omit<UserProfile, 'id' | 'created_at' | 'updated_at'>>
      }
      data_models: {
        Row: DataModel
        Insert: Omit<DataModel, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<DataModel, 'id' | 'created_at' | 'updated_at'>>
      }
      field_types: {
        Row: FieldType
        Insert: Omit<FieldType, 'id'>
        Update: Partial<Omit<FieldType, 'id'>>
      }
      fields: {
        Row: Field
        Insert: Omit<Field, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<Field, 'id' | 'created_at' | 'updated_at'>>
      }
      permissions: {
        Row: Permission
        Insert: Omit<Permission, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<Permission, 'id' | 'created_at' | 'updated_at'>>
      }
      audit_logs: {
        Row: AuditLog
        Insert: Omit<AuditLog, 'id' | 'created_at'>
        Update: never
      }
      servers: {
        Row: Server
        Insert: Omit<Server, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<Server, 'id' | 'created_at' | 'updated_at'>>
      }
      ssl_certificates: {
        Row: SSLCertificate
        Insert: Omit<SSLCertificate, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<SSLCertificate, 'id' | 'created_at' | 'updated_at'>>
      }
      applications: {
        Row: Application
        Insert: Omit<Application, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<Application, 'id' | 'created_at' | 'updated_at'>>
      }
      services: {
        Row: Service
        Insert: Omit<Service, 'id' | 'created_at' | 'updated_at'>
        Update: Partial<Omit<Service, 'id' | 'created_at' | 'updated_at'>>
      }
    }
    Views: {}
    Functions: {}
    Enums: {}
  }
}
