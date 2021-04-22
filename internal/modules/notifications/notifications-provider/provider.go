package notifications_provider

import "github.com/Confialink/wallet-accounts/internal/modules/notifications"

func Providers() []interface{} {
	return []interface{}{
		notifications.NewService,
	}
}
