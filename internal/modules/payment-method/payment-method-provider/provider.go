package payment_method_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/payment-method/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/payment-method/repository"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewPaymentMethodRepository,

		handler.NewPaymentMethodHandler,
	}
}
