package card_type_format

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-format/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-format/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-format/service"
)

func Providers() []interface{} {
	return []interface{}{
		service.NewCardTypeFormatService,
		repository.NewCardTypeFormatRepository,

		handler.NewCardTypeFormatHandler,
	}
}
