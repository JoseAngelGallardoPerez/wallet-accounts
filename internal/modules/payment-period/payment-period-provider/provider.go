package payment_period_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/payment-period/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/payment-period/repository"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewPaymentPeriodRepository,

		handler.NewPaymentPeriodHandler,
	}
}
