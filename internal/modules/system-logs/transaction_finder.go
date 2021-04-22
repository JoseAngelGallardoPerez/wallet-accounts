package system_logs

import (
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
)

type TransactionFinder struct{}

func NewTransactionFinder() *TransactionFinder {
	return &TransactionFinder{}
}

func (f *TransactionFinder) getTransactionByPurpose(
	transactions []*transactionModel.Transaction,
	purpose string,
) *transactionModel.Transaction {
	for _, v := range transactions {
		if *v.Purpose == purpose {
			return v
		}
	}
	return nil
}
