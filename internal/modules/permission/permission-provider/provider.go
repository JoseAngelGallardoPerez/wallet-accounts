package permission_provider

import "github.com/Confialink/wallet-accounts/internal/modules/permission"

func Providers() []interface{} {
	return []interface{}{
		permission.NewPermissionService,
	}
}
