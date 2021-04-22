package policy

import (
	"log"

	"github.com/Confialink/wallet-accounts/internal/modules/permission"
	"github.com/Confialink/wallet-accounts/internal/modules/policy"
	"github.com/Confialink/wallet-users/rpc/proto/users"
)

var permissionService *permission.Service

func init() {
	permissionService = permission.NewPermissionService()
}

// CheckPermission calls permission service in order to check if user granted permission
func CheckPermission(permissionValue interface{}, user *users.User) bool {
	perm := permissionValue.(permission.Permission)
	result, err := permissionService.Check(user.UID, string(perm))
	if err != nil {
		log.Printf("permission policy failed to check permission: %s", err.Error())
		return false
	}
	return result
}

func ProvideCheckSpecificPermission(permission permission.Permission) policy.Policy {
	return func(_ interface{}, user *users.User) bool {
		return CheckPermission(permission, user)
	}
}
