package handler

import (
	"fmt"

	"github.com/Confialink/wallet-pkg-list_params"

	"github.com/Confialink/wallet-accounts/internal/modules/request/filters"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/service"
)

type HandlerPrams struct {
	repository      repository.RequestRepositoryInterface
	requestsService *service.RequestsService
	requestIncludes *service.Includes
}

func NewHandlerPrams(
	repository repository.RequestRepositoryInterface,
	requestsService *service.RequestsService,
	requestIncludes *service.Includes,
) *HandlerPrams {
	return &HandlerPrams{
		repository:      repository,
		requestsService: requestsService,
		requestIncludes: requestIncludes,
	}
}

var listSelectedFields = []interface{}{
	"Id", "UserId", "Status", "Subject", "BaseCurrencyCode", "Amount", "Input", "Description", "CreatedAt",
	"StatusChangedAt", "ReferenceCurrencyCode",
	map[string][]interface{}{"User": {"Id", "Username", "Email", "FirstName", "LastName"}},
}

var userFields = []string{
	"UID",
	"Username",
	"IsCorporate",
	"CompanyDetails.CompanyName",
	"FirstName",
	"LastName",
	"UserGroupId",
	"UserGroup.Name",
	"CompanyID",
}

func (p *HandlerPrams) forAdmin(query string) *list_params.ListParams {
	listParams := list_params.NewListParamsFromQuery(query, model.Request{})
	listParams.AllowSelectFields(listSelectedFields)
	listParams.AllowPagination()
	p.addIncludes(listParams)
	p.addFilters(listParams)
	p.addSortings(listParams)
	return listParams
}

func (p *HandlerPrams) forClient(query string) *list_params.ListParams {
	listParams := list_params.NewListParamsFromQuery(query, model.Request{})
	listParams.AllowSelectFields(listSelectedFields)
	listParams.AllowPagination()
	p.addIncludes(listParams)
	p.addFilters(listParams)
	p.addSortings(listParams)
	return listParams
}

func (p *HandlerPrams) adminCsv(query string) *list_params.ListParams {
	params := p.forAdmin(query)
	params.Pagination.PageSize = 0
	params.AllowIncludes([]string{
		"user",
		"transactions",
		"DataOwt.BankDetails",
		"DataOwt.BankDetails.Country",
		"DataOwt.BeneficiaryCustomer",
		"DataOwt.IntermediaryBankDetails",
		"DataOwt.IntermediaryBankDetails.Country",
		"DataOwt.Fee",
	})
	params.Includes.AddIncludes("user")
	params.Includes.AddIncludes("transactions")
	params.Includes.AddIncludes("DataOwt.BankDetails")
	params.Includes.AddIncludes("DataOwt.BankDetails.Country")
	params.Includes.AddIncludes("DataOwt.BeneficiaryCustomer")
	params.Includes.AddIncludes("DataOwt.IntermediaryBankDetails")
	params.Includes.AddIncludes("DataOwt.IntermediaryBankDetails.Country")
	params.Includes.AddIncludes("DataOwt.Fee")
	params.Includes.AddCustomIncludes("user", func(records []interface{}) error {
		return p.requestsService.LoadFullUsers(getRequests(records), userFields)
	})
	return params
}

/*func (p *HandlerPrams) userCsv(query string) *list_params.ListParams {
	params := p.forClient(query)
	params.Pagination.PageSize = 0
	params.AllowIncludes([]string{"user", "balanceSnapshots", "balanceSnapshots.balanceType",
		"balanceDifference", "transactions"})
	params.Includes.AddIncludes("balanceSnapshots")
	params.Includes.AddIncludes("balanceSnapshots.balanceType")
	params.Includes.AddIncludes("balanceDifference")
	params.Includes.AddIncludes("transactions")
	params.Includes.AddCustomIncludes("user", func(records []interface{}) error {
		return p.requestsService.LoadFullUsers(getRequests(records), userFields)
	})
	return params
}*/

func (p *HandlerPrams) addIncludes(params *list_params.ListParams) {
	params.AllowIncludes([]string{
		"user",
		"balanceSnapshots",
		"balanceSnapshots.balanceType",
		"balanceDifference",
		"transactions",
	})

	params.Includes.AddIncludes("transactions")
	params.AddCustomIncludes("user", p.userIncludes)
	params.AddCustomIncludes("balanceDifference", p.requestIncludes.BalanceDifference)
}

func (p *HandlerPrams) userIncludes(records []interface{}) error {
	requests := make([]*model.Request, len(records))
	for i, v := range records {
		requests[i] = v.(*model.Request)
	}
	return p.repository.FillUsers(requests)
}

func (p *HandlerPrams) addFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{
		"id",
		"baseCurrencyCode",
		"subject",
		"status",
		list_params.FilterLte("createdAt"),
		list_params.FilterGte("createdAt"),
		list_params.FilterLte("statusChangedAt"),
		list_params.FilterGte("statusChangedAt"),
		list_params.FilterNin("status"),
		"isVisible",
		"accountId",
	})
	params.AddCustomFilter("id", filters.RequestIdEq)
	params.AddCustomFilter("isVisible", list_params.BoolFilter("is_visible"))
	params.AddCustomFilter("accountId", filters.AccountIdEq)
}

func (p *HandlerPrams) addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"id", "createdAt", "statusChangedAt", "user.username", "amount",
		"subject", "status", "totalOutgoingAmount", "description"})
	params.AddCustomSortings("user.username", p.userUsernameSorting)
	params.AddCustomSortings("totalOutgoingAmount", sortByTotalOutgoingAmount)
}

func sortByTotalOutgoingAmount(direction string, params *list_params.ListParams) (string, error) {
	params.AddLeftJoin("transactions neg_tx", "neg_tx.request_id = requests.id and neg_tx.amount < 0")
	params.SetGroupBy("requests.id")
	return fmt.Sprintf("ABS(SUM(neg_tx.amount)) %s", direction), nil
}

//TODO: make sorting by username
func (p *HandlerPrams) userUsernameSorting(direction string,
	params *list_params.ListParams) (string, error) {
	return "", nil
}

func getRequests(records []interface{}) []*model.Request {
	requests := make([]*model.Request, len(records))
	for i, v := range records {
		requests[i] = v.(*model.Request)
	}

	return requests
}
