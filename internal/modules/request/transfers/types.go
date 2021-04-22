package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
)

// TransactionsContainer declares that instance that contains transactions
type TransactionsContainer interface {
	// Transactions returns a list of transactions in the order that they were created
	Transactions() []*model.Transaction
}

type transactionsContainer struct {
	transactions []*model.Transaction
}

func (t *transactionsContainer) Transactions() []*model.Transaction {
	if t.transactions == nil {
		return make([]*model.Transaction, 0)
	}
	return t.transactions
}

func (t *transactionsContainer) appendTransaction(transaction *model.Transaction) {
	if t.transactions == nil {
		t.transactions = []*model.Transaction{transaction}
		return
	}
	t.transactions = append(t.transactions, transaction)
}
