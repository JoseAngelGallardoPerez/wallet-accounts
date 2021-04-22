package builders

import (
	"strconv"

	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"

	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
)

// ReceiveBuilder implements Builder interface.
type ReceiveBuilder struct {
	file   *csv.File
	helper *Helper
}

func NewReceiveBuilder(file *csv.File) *ReceiveBuilder {
	file.Name = "receive-transactions-history.csv"
	return &ReceiveBuilder{file, &Helper{}}
}

func (b *ReceiveBuilder) MakeHeader() {
	header := []string{
		"Status",
		"Currency",
		"Amount",
		"Total Amount",
		"Date Created",
		"Date Processed",
		"Source Address",
		"Destination Address",
	}
	b.file.WriteRow(header)
}

func (b *ReceiveBuilder) MakeBody(
	items []*transactionModel.Transaction,
	requestsMap map[uint64]interface{},
) {
	timeSettings, _ := syssettings.GetTimeSettings()

	for _, v := range items {
		requestData, _ := requestsMap[*v.RequestId].(map[string]interface{})

		transactions := requestData["transactions"].([]*transactionModel.Transaction)
		request := requestData["request"].(*requestModel.Request)

		amount := *v.Amount
		if v.ShowAmount != nil {
			amount = *v.ShowAmount
		}

		formattedCreatedDate := timefmt.Format(*v.CreatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)
		formattedUpdatedDate := timefmt.Format(*v.UpdatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)

		source := b.helper.getOutgoingTransaction(transactions)
		target := b.helper.getIncomingTransaction(transactions)

		record := []string{
			*v.Status,
			*request.BaseCurrencyCode,
			amount.String(),
			amount.String(),
			formattedCreatedDate,
			formattedUpdatedDate,
			strconv.FormatUint(*source.Id, 10),
			strconv.FormatUint(*target.Id, 10),
		}

		b.file.WriteRow(record)
	}

	b.file.WriteRow([]string{})
}
