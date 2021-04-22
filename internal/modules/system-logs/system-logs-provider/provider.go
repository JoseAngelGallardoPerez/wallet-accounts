package system_logs_provider

import system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"

func Providers() []interface{} {
	return []interface{}{
		system_logs.NewTransactionFinder,
		system_logs.NewSystemLogsService,
		system_logs.NewTransactionLogCreator,
		system_logs.NewIwtBankDetailsLogCreator,
		system_logs.NewAccountLogCreator,
		system_logs.NewAccountTypeLogCreator,
		system_logs.NewCardLogCreator,
		system_logs.NewCardTypeLogCreator,
	}
}
