export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[]

export type Database = {
  public: {
    Tables: {
      profiles: {
        Row: {
          id: string
          email: string
          full_name: string | null
          avatar_url: string | null
          role: 'admin' | 'maintainer' | 'viewer'
          created_at: string
          updated_at: string
        }
        Insert: {
          id: string
          email: string
          full_name?: string | null
          avatar_url?: string | null
          role?: 'admin' | 'maintainer' | 'viewer'
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          email?: string
          full_name?: string | null
          avatar_url?: string | null
          role?: 'admin' | 'maintainer' | 'viewer'
          created_at?: string
          updated_at?: string
        }
      }
      roles: {
        Row: {
          id: string
          name: string
          description: string | null
          created_at: string
        }
        Insert: {
          id?: string
          name: string
          description?: string | null
          created_at?: string
        }
        Update: {
          id?: string
          name?: string
          description?: string | null
          created_at?: string
        }
      }
      permissions: {
        Row: {
          id: string
          role_id: string
          entity_type_id: string | null
          can_create: boolean
          can_read: boolean
          can_update: boolean
          can_delete: boolean
          created_at: string
        }
        Insert: {
          id?: string
          role_id: string
          entity_type_id?: string | null
          can_create?: boolean
          can_read?: boolean
          can_update?: boolean
          can_delete?: boolean
          created_at?: string
        }
        Update: {
          id?: string
          role_id?: string
          entity_type_id?: string | null
          can_create?: boolean
          can_read?: boolean
          can_update?: boolean
          can_delete?: boolean
          created_at?: string
        }
      }
      entities: {
        Row: {
          id: string
          name: string
          display_name: string
          description: string | null
          icon: string | null
          created_by: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          name: string
          display_name: string
          description?: string | null
          icon?: string | null
          created_by?: string | null
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          name?: string
          display_name?: string
          description?: string | null
          icon?: string | null
          created_by?: string | null
          created_at?: string
          updated_at?: string
        }
      }
      attributes: {
        Row: {
          id: string
          entity_id: string
          name: string
          display_name: string
          attribute_type: 'string' | 'number' | 'date' | 'relation'
          is_required: boolean
          default_value: string | null
          references_entity_id: string | null
          validation_rules: Json | null
          order_index: number
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          entity_id: string
          name: string
          display_name: string
          attribute_type: 'string' | 'number' | 'date' | 'relation'
          is_required?: boolean
          default_value?: string | null
          references_entity_id?: string | null
          validation_rules?: Json | null
          order_index?: number
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          entity_id?: string
          name?: string
          display_name?: string
          attribute_type?: 'string' | 'number' | 'date' | 'relation'
          is_required?: boolean
          default_value?: string | null
          references_entity_id?: string | null
          validation_rules?: Json | null
          order_index?: number
          created_at?: string
          updated_at?: string
        }
      }
      records: {
        Row: {
          id: string
          entity_id: string
          created_by: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          entity_id: string
          created_by?: string | null
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          entity_id?: string
          created_by?: string | null
          created_at?: string
          updated_at?: string
        }
      }
      values: {
        Row: {
          id: string
          record_id: string
          attribute_id: string
          value_text: string | null
          value_number: number | null
          value_date: string | null
          value_relation: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          record_id: string
          attribute_id: string
          value_text?: string | null
          value_number?: number | null
          value_date?: string | null
          value_relation?: string | null
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          record_id?: string
          attribute_id?: string
          value_text?: string | null
          value_number?: number | null
          value_date?: string | null
          value_relation?: string | null
          created_at?: string
          updated_at?: string
        }
      }
      audit_logs: {
        Row: {
          id: string
          user_id: string | null
          action: 'create' | 'update' | 'delete' | 'view'
          entity_type: string
          record_id: string | null
          old_values: Json | null
          new_values: Json | null
          ip_address: string | null
          user_agent: string | null
          created_at: string
        }
        Insert: {
          id?: string
          user_id?: string | null
          action: 'create' | 'update' | 'delete' | 'view'
          entity_type: string
          record_id?: string | null
          old_values?: Json | null
          new_values?: Json | null
          ip_address?: string | null
          user_agent?: string | null
          created_at?: string
        }
        Update: {
          id?: string
          user_id?: string | null
          action?: 'create' | 'update' | 'delete' | 'view'
          entity_type?: string
          record_id?: string | null
          old_values?: Json | null
          new_values?: Json | null
          ip_address?: string | null
          user_agent?: string | null
          created_at?: string
        }
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      get_user_role: {
        Args: { user_uuid: string }
        Returns: 'admin' | 'maintainer' | 'viewer'
      }
      has_permission: {
        Args: {
          user_uuid: string
          entity_uuid: string
          permission_type: string
        }
        Returns: boolean
      }
      check_record_references: {
        Args: { record_uuid: string }
        Returns: boolean
      }
    }
    Enums: {
      user_role: 'admin' | 'maintainer' | 'viewer'
      attribute_type: 'string' | 'number' | 'date' | 'relation'
      audit_action: 'create' | 'update' | 'delete' | 'view'
    }
  }
}
