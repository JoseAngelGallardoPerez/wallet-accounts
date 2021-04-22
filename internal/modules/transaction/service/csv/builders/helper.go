package builders

import (
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/shopspring/decimal"
)

type Helper struct{}

func (h *Helper) getFees(transactions []*transactionModel.Transaction) []*transactionModel.Transaction {
	fees := make([]*transactionModel.Transaction, 0, 1)
	for _, item := range transactions {
		if item.IsExchangeMarginFee() || item.IsDefaultTransferFee() {
			fees = append(fees, item)
		}
	}
	return fees
}

func (h *Helper) getOutgoingTransaction(transactions []*transactionModel.Transaction) *transactionModel.Transaction {
	for _, item := range transactions {
		if item.IsTargetOutgoing() {
			return item
		}
	}
	return nil
}

func (h *Helper) getIncomingTransaction(transactions []*transactionModel.Transaction) *transactionModel.Transaction {
	for _, item := range transactions {
		if item.IsIncoming() {
			return item
		}
	}
	return nil
}

func (h *Helper) calculateTotalFees(fees []*transactionModel.Transaction) decimal.Decimal {
	var totalFee decimal.Decimal
	for _, fee := range fees {
		totalFee = totalFee.Add(*fee.Amount)
	}
	return totalFee
}
