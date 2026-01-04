import { Database } from '@/types/database';

export type UserRole = Database['public']['Enums']['user_role'];
export type PermissionType = 'create' | 'read' | 'update' | 'delete';

export interface Permission {
  can_create: boolean;
  can_read: boolean;
  can_update: boolean;
  can_delete: boolean;
}

export interface UserPermissions {
  role: UserRole;
  permissions: Record<string, Permission>;
}

/**
 * Check if a role has a specific permission
 */
export function hasRolePermission(
  role: UserRole,
  permission: PermissionType
): boolean {
  const rolePermissions: Record<UserRole, Permission> = {
    admin: {
      can_create: true,
      can_read: true,
      can_update: true,
      can_delete: true,
    },
    maintainer: {
      can_create: true,
      can_read: true,
      can_update: true,
      can_delete: false,
    },
    viewer: {
      can_create: false,
      can_read: true,
      can_update: false,
      can_delete: false,
    },
  };

  const permissions = rolePermissions[role];
  return permissions[`can_${permission}`];
}

/**
 * Check if a user can perform an action on an entity
 */
export function canPerformAction(
  userRole: UserRole,
  action: PermissionType,
  entityPermissions?: Permission
): boolean {
  // Admin always has full access
  if (userRole === 'admin') {
    return true;
  }

  // If entity-specific permissions are provided, use those
  if (entityPermissions) {
    return entityPermissions[`can_${action}`];
  }

  // Otherwise, use role-based permissions
  return hasRolePermission(userRole, action);
}

/**
 * Get all permissions for a role
 */
export function getRolePermissions(role: UserRole): Permission {
  const rolePermissions: Record<UserRole, Permission> = {
    admin: {
      can_create: true,
      can_read: true,
      can_update: true,
      can_delete: true,
    },
    maintainer: {
      can_create: true,
      can_read: true,
      can_update: true,
      can_delete: false,
    },
    viewer: {
      can_create: false,
      can_read: true,
      can_update: false,
      can_delete: false,
    },
  };

  return rolePermissions[role];
}

/**
 * Check if a record can be deleted (no references)
 */
export function canDeleteRecord(
  userRole: UserRole,
  hasReferences: boolean
): { canDelete: boolean; reason?: string } {
  if (!hasRolePermission(userRole, 'delete')) {
    return {
      canDelete: false,
      reason: 'You do not have permission to delete records',
    };
  }

  if (hasReferences) {
    return {
      canDelete: false,
      reason: 'Cannot delete record that is referenced by other records',
    };
  }

  return { canDelete: true };
}
