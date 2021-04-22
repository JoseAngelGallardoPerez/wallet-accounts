package card_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card/http/handlers"
	"github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/card/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/card/service"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewCardRepository,
		service.NewCardService,
		service.NewCsv,
		serializer.NewCardSerializer,

		handlers.NewHandlerParams,
		handlers.NewCardHandler,
		handlers.NewCsvHandler,
		handlers.NewCardListHandler,
	}
}
