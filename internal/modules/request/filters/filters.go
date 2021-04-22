package filters

import "github.com/Confialink/wallet-pkg-list_params"

func UserIdEq(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	params.AddInnerJoin("transactions", "requests.id = transactions.request_id")
	params.AddLeftJoin("accounts", "accounts.id = transactions.account_id")
	params.AddLeftJoin("cards", "cards.id = transactions.card_id")

	return "(accounts.user_id = ? OR cards.user_id = ?)", []string{inputValues[0], inputValues[0]}
}

func IncomingByStatus(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	params.AddInnerJoin("transactions", "requests.id = transactions.request_id")
	return "((transactions.status = ? AND transactions.amount >= 0) " +
		"OR transactions.amount < 0)", inputValues[0]
}

func AccountIdEq(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	params.AddInnerJoin("transactions", "requests.id = transactions.request_id")
	params.AddLeftJoin("accounts", "accounts.id = transactions.account_id")
	return "accounts.id = ?", inputValues[0]
}

func CardIdEq(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	params.AddInnerJoin("transactions", "requests.id = transactions.request_id")
	params.AddLeftJoin("cards", "cards.id = transactions.card_id")
	return "cards.id = ?", inputValues[0]
}

func RequestIdEq(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	return "requests.id LIKE ?", inputValues[0] + "%"
}
