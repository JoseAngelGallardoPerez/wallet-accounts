package handler

import (
	"github.com/Confialink/wallet-pkg-list_params"
	"fmt"
)

func IncomingByStatus(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	return "((transactions.status = ? AND transactions.amount >= 0) " +
		"OR transactions.amount < 0)", inputValues[0]
}

const (
	OperationsOutgoing = "outgoing"
	OperationsIncoming = "incoming"
)

func operationFilter(inputValues []string, params *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	separator := "0"

	if inputValues[0] == OperationsOutgoing {
		return "transactions.amount < ?", separator
	} else if inputValues[0] == OperationsIncoming {
		return "transactions.amount >= ?", separator
	}

	return "1 = ?", "1" // TODO: add ability to ignore filter into list_params
}

func searchFilter(inputValues []string, params *list_params.ListParams,
) (dbConditionPart string, dbValues interface{}) {
	condition := "transactions.description LIKE ?"
	return condition, fmt.Sprintf("%%%s%%", inputValues[0])
}
