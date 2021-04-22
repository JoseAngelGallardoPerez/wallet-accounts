package filters

import "github.com/Confialink/wallet-pkg-list_params"

func UserGroupEq(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	params.AddInnerJoin("transfer_fees_user_groups", "transfer_fees_user_groups.transfer_fee_id = transfer_fees.id")
	return "(transfer_fees_user_groups.user_group_id = ?)", inputValues[0]
}

func CurrencyCodeEq(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	params.AddInnerJoin("transfer_fees_parameters", "transfer_fees_parameters.transfer_fee_id = transfer_fees.id")
	return "(transfer_fees_parameters.currency_code = ?)", inputValues[0]
}
