package request

import "github.com/Confialink/wallet-accounts/internal/modules/settings"

const (
	//Transfer Between Accounts
	SettingTbaActionRequired = settings.Name("tba_action_required")
	SettingTbaTanRequired    = settings.Name("tba_tan_required")
	//Transfer Between Users
	SettingTbuActionRequired = settings.Name("tbu_action_required")
	SettingTbuTanRequired    = settings.Name("tbu_tan_required")
	//Outgoing Wire Transfer
	SettingOwtActionRequired = settings.Name("owt_action_required")
	SettingOwtTanRequired    = settings.Name("owt_tan_required")
	//Card Funding Transfer
	SettingCftActionRequired = settings.Name("cft_action_required")
	SettingCftTanRequired    = settings.Name("cft_tan_required")
	//CreditFromAlias Account
	SettingCreditAccountActionRequired = settings.Name("credit_account_action_required")
)
