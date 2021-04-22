package app

import (
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/app/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/app/validator"
)

func Providers() []interface{} {
	return []interface{}{
		//AuthHandlerServiceInterface
		service.NewContext,
		serializer.NewModelSerializer,
		validator.NewValidator,

		handler.NewNotFoundHandler,
		handler.NewCorsHandler,
	}
}
