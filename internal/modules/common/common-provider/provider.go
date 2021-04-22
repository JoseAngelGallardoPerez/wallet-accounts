package common_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/common/http/handlers"
	"github.com/Confialink/wallet-accounts/internal/modules/common/service/forms/service"
)

func Providers() []interface{} {
	return []interface{}{
		service.NewModelFormService,
		handlers.NewModelFormHandler,
	}
}
