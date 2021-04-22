package builders

import (
	"strconv"

	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"

	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"
)

// BuyBuilder implements Builder interface.
type BuyBuilder struct {
	file   *csv.File
	helper *Helper
}

func NewBuyBuilder(file *csv.File) *BuyBuilder {
	file.Name = "buy-transactions-history.csv"
	return &BuyBuilder{file, &Helper{}}
}

func (b *BuyBuilder) MakeHeader() {
	header := []string{
		"Status",
		"Number",
		"Method",
		"Source Amount",
		"Amount/EUR",
		"Target Amount",
		"Promo code",
		"Transaction Fee",
		"Total Amount",
		"Date Submitted",
		"Date Processed",
	}
	b.file.WriteRow(header)
}

func (b *BuyBuilder) MakeBody(
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
			strconv.FormatUint(*v.RequestId, 10),
			*v.Type,
			source.Amount.String(),
			amount.String(),
			target.Amount.String(),
			"",
			totalFee.String(),
			formattedCreatedDate,
			formattedUpdatedDate,
		}

		b.file.WriteRow(record)
	}

	b.file.WriteRow([]string{})
}
