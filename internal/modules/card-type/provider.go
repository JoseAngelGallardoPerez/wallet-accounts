package card_type

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/service"
)

func Providers() []interface{} {
	return []interface{}{
		serializer.NewCardTypeSerializer,
		service.NewCardTypeService,
		repository.NewCardTypeRepository,

		handler.NewCardTypeHandler,
	}
}
