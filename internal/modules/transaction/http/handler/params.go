package handler

import (
	"fmt"

	"github.com/Confialink/wallet-pkg-list_params"

	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
)

var (
	historyOutputFields = []interface{}{
		"Id", "RequestId", "Status", "Type", "Amount", "Purpose", "IsVisible", "AvailableBalanceSnapshot", "Description",
		"CreatedAt", "UpdatedAt",
	}
)

type HandlerParams struct {
}

func NewHandlerParams() *HandlerParams {
	return &HandlerParams{}
}

func (p *HandlerParams) forUser(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.Transaction{})
	params.AllowPagination()
	p.addFilters(params)
	p.addSortings(params)
	return params
}

func (p *HandlerParams) forUserCsv(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.Transaction{})
	params.Pagination.PageSize = 0
	p.addFilters(params)
	p.addSortings(params)
	return params
}

func (p *HandlerParams) forClientHistory(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.Transaction{})
	params.AllowSelectFields(historyOutputFields)
	params.AllowPagination()

	params.AllowFilters([]string{
		"status", "purpose", "type", "createdAtTo", "createdAtFrom", "updatedAtTo", "updatedAtFrom", "operation", "q",
	})

	params.AddCustomFilter("createdAtFrom", list_params.DateFromFilter("transactions.created_at"))
	params.AddCustomFilter("createdAtTo", list_params.DateToFilter("transactions.created_at"))

	params.AddCustomFilter("updatedAtFrom", list_params.DateFromFilter("transactions.updated_at"))
	params.AddCustomFilter("updatedAtTo", list_params.DateToFilter("transactions.updated_at"))

	params.AddCustomFilter("operation", operationFilter)
	params.AddCustomFilter("q", searchFilter)

	params.AllowSortings([]string{
		"status", "amount", "createdAt", "updatedAt",
	})

	return params
}

func (p *HandlerParams) forClientHistoryCsv(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.Transaction{})
	params.AllowSelectFields(historyOutputFields)
	params.Pagination.PageSize = 0

	params.AllowFilters([]string{
		"status", "purpose", "type", "createdAtTo", "createdAtFrom", "updatedAtTo", "updatedAtFrom",
	})

	return params
}

func (p *HandlerParams) addFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{
		"id",
		"createdAtTo",
		"createdAtFrom",
		"accountId",
		"type",
		"is_visible",
		"incomingByStatus",
		"accounts.user_id",
		list_params.FilterIn("id"),
	})
	params.AddCustomFilter("createdAtFrom", list_params.DateFromFilter("transactions.created_at"))
	params.AddCustomFilter("createdAtTo", list_params.DateToFilter("transactions.created_at"))
	params.AddCustomFilter("incomingByStatus", IncomingByStatus)
}

func (p *HandlerParams) addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"id", "createdAt", "description", "amount", "status", "statusChangedAt"})
	params.AddCustomSortings("statusChangedAt", transactionStatusChangedAt)
}

func transactionStatusChangedAt(direction string,
	params *list_params.ListParams) (string, error) {
	params.AddLeftJoin("requests", "transactions.request_id = requests.id")
	// order by request status changed at and transaction id
	return fmt.Sprintf("requests.status_changed_at %s, transactions.id DESC", direction), nil
}
