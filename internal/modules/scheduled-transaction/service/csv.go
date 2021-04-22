package service

import (
	"fmt"
	"strconv"
	"time"

	scheduled_transaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"

	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/Confialink/wallet-pkg-utils/timefmt"

	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
)

func GetCsv(transactions []*scheduled_transaction.ScheduledTransaction) (*csv.File, error) {
	currentTime := time.Now()
	timeSettings, err := syssettings.GetTimeSettings()
	if err != nil {
		return nil, err
	}

	file := csv.NewFile()
	formattedCurrentTime := timefmt.FormatFilenameWithTime(currentTime, timeSettings.Timezone)
	file.Name = fmt.Sprintf("scheduled-transactions-%s.csv", formattedCurrentTime)

	writeTransactions(transactions, timeSettings, file)
	return file, nil
}

func writeTransactions(transactions []*scheduled_transaction.ScheduledTransaction, timeSettings *syssettings.TimeSettings, file *csv.File) {
	transactionsHeader := []string{"Id", "Task", "Account number", "Amount", "Status", "Scheduled date", "Created date", "Updated date"}
	file.WriteRow(transactionsHeader)

	for _, v := range transactions {

		formattedScheduledDate := timefmt.Format(*v.ScheduledDate, timeSettings.DateTimeFormat, timeSettings.Timezone)
		formattedCreatedDate := timefmt.Format(*v.CreatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)
		formattedUpdateDate := timefmt.Format(*v.UpdatedAt, timeSettings.DateTimeFormat, timeSettings.Timezone)

		record := []string{
			strconv.FormatUint(*v.Id, 10),
			v.Reason.Description(),
			v.Account.Number,
			v.Amount.String(),
			string(v.Status),
			formattedScheduledDate,
			formattedCreatedDate,
			formattedUpdateDate,
		}
		file.WriteRow(record)
	}
}
