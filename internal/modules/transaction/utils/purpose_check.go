package utils

import (
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
)

func IsMainTransactionPurpose(purpose constants.Purpose) bool {
	for _, v := range constants.MainTransactions {
		if purpose == v {
			return true
		}
	}
	return false
}
