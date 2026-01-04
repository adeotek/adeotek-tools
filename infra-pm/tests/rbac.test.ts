import { describe, it, expect } from 'vitest';
import {
  hasRolePermission,
  canPerformAction,
  getRolePermissions,
  canDeleteRecord,
} from '../lib/rbac/permissions';

describe('RBAC Permissions', () => {
  describe('hasRolePermission', () => {
    it('should allow admin to perform all actions', () => {
      expect(hasRolePermission('admin', 'create')).toBe(true);
      expect(hasRolePermission('admin', 'read')).toBe(true);
      expect(hasRolePermission('admin', 'update')).toBe(true);
      expect(hasRolePermission('admin', 'delete')).toBe(true);
    });

    it('should allow maintainer to create, read, and update but not delete', () => {
      expect(hasRolePermission('maintainer', 'create')).toBe(true);
      expect(hasRolePermission('maintainer', 'read')).toBe(true);
      expect(hasRolePermission('maintainer', 'update')).toBe(true);
      expect(hasRolePermission('maintainer', 'delete')).toBe(false);
    });

    it('should only allow viewer to read', () => {
      expect(hasRolePermission('viewer', 'create')).toBe(false);
      expect(hasRolePermission('viewer', 'read')).toBe(true);
      expect(hasRolePermission('viewer', 'update')).toBe(false);
      expect(hasRolePermission('viewer', 'delete')).toBe(false);
    });
  });

  describe('canPerformAction', () => {
    it('should always allow admin regardless of entity permissions', () => {
      const restrictedPermissions = {
        can_create: false,
        can_read: false,
        can_update: false,
        can_delete: false,
      };

      expect(canPerformAction('admin', 'create', restrictedPermissions)).toBe(
        true
      );
      expect(canPerformAction('admin', 'delete', restrictedPermissions)).toBe(
        true
      );
    });

    it('should respect entity-specific permissions for non-admin roles', () => {
      const customPermissions = {
        can_create: true,
        can_read: true,
        can_update: false,
        can_delete: false,
      };

      expect(canPerformAction('maintainer', 'create', customPermissions)).toBe(
        true
      );
      expect(canPerformAction('maintainer', 'update', customPermissions)).toBe(
        false
      );
    });

    it('should use role-based permissions when no entity permissions provided', () => {
      expect(canPerformAction('maintainer', 'create')).toBe(true);
      expect(canPerformAction('maintainer', 'delete')).toBe(false);
      expect(canPerformAction('viewer', 'update')).toBe(false);
    });
  });

  describe('getRolePermissions', () => {
    it('should return correct permissions for each role', () => {
      const adminPerms = getRolePermissions('admin');
      expect(adminPerms.can_create).toBe(true);
      expect(adminPerms.can_delete).toBe(true);

      const maintainerPerms = getRolePermissions('maintainer');
      expect(maintainerPerms.can_create).toBe(true);
      expect(maintainerPerms.can_delete).toBe(false);

      const viewerPerms = getRolePermissions('viewer');
      expect(viewerPerms.can_create).toBe(false);
      expect(viewerPerms.can_read).toBe(true);
    });
  });

  describe('canDeleteRecord', () => {
    it('should prevent deletion if user lacks delete permission', () => {
      const result = canDeleteRecord('viewer', false);
      expect(result.canDelete).toBe(false);
      expect(result.reason).toContain('permission');
    });

    it('should prevent deletion if record has references', () => {
      const result = canDeleteRecord('admin', true);
      expect(result.canDelete).toBe(false);
      expect(result.reason).toContain('referenced');
    });

    it('should allow deletion for admin with no references', () => {
      const result = canDeleteRecord('admin', false);
      expect(result.canDelete).toBe(true);
      expect(result.reason).toBeUndefined();
    });

    it('should prevent maintainer from deleting even without references', () => {
      const result = canDeleteRecord('maintainer', false);
      expect(result.canDelete).toBe(false);
      expect(result.reason).toContain('permission');
    });
  });
});
