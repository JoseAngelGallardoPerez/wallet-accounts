package handler

import (
	"fmt"

	scheduled_transaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"

	"github.com/Confialink/wallet-pkg-list_params"
)

var showListOutputFields = []interface{}{
	"Id",
	"Amount",
	"Reason",
	"Status",
	"ScheduledDate",
	"CreatedAt",
	"UpdatedAt",
	map[string][]interface{}{"Account": {
		"Number",
		map[string][]interface{}{"Type": {
			"CurrencyCode",
		}},
	}},
}

func getListParams(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, scheduled_transaction.ScheduledTransaction{})
	params.AllowSelectFields(showListOutputFields)
	params.AllowPagination()
	addIncludes(params)
	addSortings(params)
	allowFilters(params)
	addFilters(params)
	return params
}

func getCsvParams(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, scheduled_transaction.ScheduledTransaction{})
	params.AllowSelectFields(showListOutputFields)
	params.Pagination.PageSize = 0
	addIncludes(params)
	addSortings(params)
	allowFilters(params)
	addFilters(params)
	return params
}

func addIncludes(params *list_params.ListParams) {
	params.AllowIncludes([]string{"account", "account.type"})
}

func addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"id", "reason", "amount", "status", "scheduledDate", "createdAt", "updatedAt"})
	params.AddCustomSortings("account.number", accountNumberSorting)
}

func accountNumberSorting(direction string,
	params *list_params.ListParams) (string, error) {
	params.AddLeftJoin("accounts", "transactions.account_id = account.id")
	return fmt.Sprintf("account.number %s", direction), nil
}

func allowFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{
		"createdAt",
		"status",
		"reason",
		list_params.FilterIn("status"),
		list_params.FilterIn("reason"),
		list_params.FilterLte("scheduledDate"),
		list_params.FilterGte("scheduledDate"),
		list_params.FilterLte("createdAt"),
		list_params.FilterGte("createdAt"),
	})
}

func addFilters(params *list_params.ListParams) {}
