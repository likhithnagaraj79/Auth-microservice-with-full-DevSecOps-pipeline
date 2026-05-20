package rbac_test

import (
	"testing"

	"github.com/likhithnagaraj79/auth-service/internal/models"
	"github.com/likhithnagaraj79/auth-service/internal/rbac"
	"github.com/stretchr/testify/assert"
)

func TestAdminHasAllPermissions(t *testing.T) {
	perms := []rbac.Permission{
		rbac.PermReadUsers,
		rbac.PermWriteUsers,
		rbac.PermDeleteUsers,
		rbac.PermReadLogs,
		rbac.PermManageRoles,
	}
	for _, p := range perms {
		assert.True(t, rbac.HasPermission(models.RoleAdmin, p), "admin should have %s", p)
	}
}

func TestUserLimitedPermissions(t *testing.T) {
	assert.True(t, rbac.HasPermission(models.RoleUser, rbac.PermReadUsers))
	assert.False(t, rbac.HasPermission(models.RoleUser, rbac.PermDeleteUsers))
	assert.False(t, rbac.HasPermission(models.RoleUser, rbac.PermManageRoles))
	assert.False(t, rbac.HasPermission(models.RoleUser, rbac.PermReadLogs))
}

func TestViewerLimitedPermissions(t *testing.T) {
	assert.True(t, rbac.HasPermission(models.RoleViewer, rbac.PermReadUsers))
	assert.False(t, rbac.HasPermission(models.RoleViewer, rbac.PermWriteUsers))
	assert.False(t, rbac.HasPermission(models.RoleViewer, rbac.PermDeleteUsers))
}

func TestUnknownRoleHasNoPermissions(t *testing.T) {
	assert.False(t, rbac.HasPermission("unknown", rbac.PermReadUsers))
}
