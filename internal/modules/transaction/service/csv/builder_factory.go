package csv

import (
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/service/csv/builders"
	"github.com/Confialink/wallet-pkg-utils/csv"
)

type BuilderCreator struct{}

func NewBuilderCreator() *BuilderCreator {
	return &BuilderCreator{}
}

func (c *BuilderCreator) CreateBuilder(action string, file *csv.File) Builder {
	var builder Builder

	switch action {
	case "tba_outgoing":
		builder = builders.NewSellBuilder(file)
	case "tba_incoming":
		builder = builders.NewBuyBuilder(file)
	case "tbu_outgoing":
		builder = builders.NewSendBuilder(file)
	case "tbu_incoming":
		builder = builders.NewReceiveBuilder(file)
	case "convert_outgoing":
		builder = builders.NewConvertBuilder(file)
	default:
		builder = builders.NewSellBuilder(file)
	}

	return builder
}
