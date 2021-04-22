package csv

import (
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-pkg-utils/csv"
)

type Builder interface {
	MakeHeader()
	MakeBody(items []*transactionModel.Transaction, requestsMap map[uint64]interface{})
}

type Producer struct {
	Builder Builder
}

func NewProducer(t string, file *csv.File) *Producer {
	builder := NewBuilderCreator().CreateBuilder(t, file)
	return &Producer{builder}
}

// Construct tells the builder what to do and in what order.
func (p *Producer) Construct(items []*transactionModel.Transaction, requestsMap map[uint64]interface{}) error {
	p.Builder.MakeHeader()
	p.Builder.MakeBody(items, requestsMap)
	return nil
}
