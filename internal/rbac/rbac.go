package rbac

import "github.com/likhithnagaraj79/auth-service/internal/models"

type Permission string

const (
	PermReadUsers   Permission = "users:read"
	PermWriteUsers  Permission = "users:write"
	PermDeleteUsers Permission = "users:delete"
	PermReadLogs    Permission = "logs:read"
	PermManageRoles Permission = "roles:manage"
)

var rolePermissions = map[models.Role][]Permission{
	models.RoleAdmin: {
		PermReadUsers,
		PermWriteUsers,
		PermDeleteUsers,
		PermReadLogs,
		PermManageRoles,
	},
	models.RoleUser: {
		PermReadUsers,
	},
	models.RoleViewer: {
		PermReadUsers,
	},
}

func HasPermission(role models.Role, perm Permission) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

func GetPermissions(role models.Role) []Permission {
	return rolePermissions[role]
}
