package tan_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/tan"
	"github.com/Confialink/wallet-accounts/internal/modules/tan/handler"
)

func Providers() []interface{} {
	return []interface{}{
		tan.NewRepository,
		tan.NewSubscriberRepository,
		tan.NewService,
		tan.NewWatcher,
		tan.NewBcryptHasherVerifier,

		handler.NewController,
	}
}
