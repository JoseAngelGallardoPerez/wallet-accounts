package tan

import "github.com/Confialink/wallet-accounts/internal/modules/settings"

const (
	SettingTanGenerateQtyInt64        = settings.Name("tan_generate_qty")
	SettingTanGenerateTriggerQtyInt64 = settings.Name("tan_generate_trigger_qty")
	SettingTanMessageSubjectString    = settings.Name("tan_message_subject")
	SettingTanMessageContentString    = settings.Name("tan_message_content")
)
