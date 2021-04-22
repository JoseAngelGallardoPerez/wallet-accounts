package fee_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/fee"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/repository"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewTransferFee,
		repository.NewTransferFeeParameters,
		fee.NewServiceTransferFee,

		handler.NewTransferFee,
	}
}
