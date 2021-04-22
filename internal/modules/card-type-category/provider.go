package card_type_category

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-category/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-category/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-category/service"
)

func Providers() []interface{} {
	return []interface{}{
		service.NewCardTypeCategoryService,
		repository.NewCardTypeCategoryRepository,

		handler.NewCardTypeCategoryHandler,
	}
}
