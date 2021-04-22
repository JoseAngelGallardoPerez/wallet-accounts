package settings_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
	"github.com/Confialink/wallet-accounts/internal/modules/settings/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/settings/repository"
)

func Providers() []interface{} {
	return []interface{}{
		settings.NewService,
		repository.NewSettings,

		handler.NewSettingsController,
	}
}
