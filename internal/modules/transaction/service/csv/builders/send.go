package builders

import (
	"strconv"

	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"

	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
)

// SendBuilder implements Builder interface.
type SendBuilder struct {
	file   *csv.File
	helper *Helper
}

func NewSendBuilder(file *csv.File) *SendBuilder {
	file.Name = "send-transactions-history.csv"
	return &SendBuilder{file, &Helper{}}
}

func (b *SendBuilder) MakeHeader() {
	header := []string{
		"Status",
		"Currency",
		"Amount",
		"Commission",
		"Transaction Fee",
		"Total Amount",
		"Date Created",
		"Date Processed",
		"Source Address",
		"Destination Address",
	}
	b.file.WriteRow(header)
}

func (b *SendBuilder) MakeBody(
	items []*transactionModel.Transaction,
	requestsMap map[uint64]interface{},
) {
	timeSettings, _ := syssettings.GetTimeSettings()

	for _, v := range items {
		requestData, _ := requestsMap[*v.RequestId].(map[string]interface{})

		transactions := requestData["transactions"].([]*transactionModel.Transaction)

		amount := *v.Amount
		if v.ShowAmount != nil {
			amount = *v.ShowAmount
		}

		formattedCreatedDate := timefmt.Format(*v.CreatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)
		formattedUpdatedDate := timefmt.Format(*v.UpdatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)

		source := b.helper.getOutgoingTransaction(transactions)
		target := b.helper.getIncomingTransaction(transactions)
		fees := b.helper.getFees(transactions)
		totalFee := b.helper.calculateTotalFees(fees)

		record := []string{
			*v.Status,
			"",
			amount.String(),
			"",
			totalFee.String(),
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
