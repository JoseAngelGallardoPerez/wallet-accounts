package builders

import (
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"

	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
)

// ConvertBuilder implements Builder interface.
type ConvertBuilder struct {
	file   *csv.File
	helper *Helper
}

func NewConvertBuilder(file *csv.File) *ConvertBuilder {
	file.Name = "convert-transactions-history.csv"
	return &ConvertBuilder{file, &Helper{}}
}

func (b *ConvertBuilder) MakeHeader() {
	header := []string{
		"Status",
		"Source Amount",
		"Target Amount",
		"Promo code",
		"Transaction Fee",
		"Total Amount",
		"Type",
		"Date Created",
		"Date Processed",
	}
	b.file.WriteRow(header)
}

func (b *ConvertBuilder) MakeBody(
	items []*transactionModel.Transaction,
	requestsMap map[uint64]interface{},
) {
	timeSettings, _ := syssettings.GetTimeSettings()

	for _, v := range items {
		requestData, _ := requestsMap[*v.RequestId].(map[string]interface{})

		transactions := requestData["transactions"].([]*transactionModel.Transaction)

		formattedCreatedDate := timefmt.Format(*v.CreatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)
		formattedUpdatedDate := timefmt.Format(*v.UpdatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)

		source := b.helper.getOutgoingTransaction(transactions)
		target := b.helper.getIncomingTransaction(transactions)
		fees := b.helper.getFees(transactions)
		totalFee := b.helper.calculateTotalFees(fees)

		record := []string{
			*v.Status,
			source.Amount.String(),
			target.Amount.String(),
			"",
			totalFee.String(),
			"",
			formattedCreatedDate,
			formattedUpdatedDate,
		}

		b.file.WriteRow(record)
	}

	b.file.WriteRow([]string{})
}
