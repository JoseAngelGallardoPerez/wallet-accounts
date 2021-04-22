package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account/service"
	userService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
)

// showListOutputFields is list of possible fields can be serialized
var showListOutputFields = []interface{}{
	"ID", "Number", "TypeID", "UserId",
	"Description", "IsActive", "Balance", "AllowWithdrawals", "AllowDeposits",
	"MaturityDate", "PayoutDay", "InterestAccountId", "AvailableAmount",
	"CreatedAt",
	map[string][]interface{}{
		"Type": {"ID", "Name", "CurrencyCode", "BalanceFeeAmount",
			"BalanceChargeDay", "BalanceLimitAmount", "CreditLimitAmount",
			"CreditAnnualInterestRate", "CreditPayoutMethodID", "CreditChargePeriodID",
			"CreditChargeDay", "DepositAnnualInterestRate", "DepositPayoutMethodID",
			"DepositPayoutPeriodID", "DepositPayoutDay", "AutoNumberGeneration",
			"MonthlyMaintenanceFee",
			map[string][]interface{}{"Currency": {"Id", "Code"}},
		},
		"User": {"UID", "Username", "FirstName", "LastName", "RoleName", "Email"},
	}}

// showListForUserOutputFields is list of possible fields can be serialized
var showListForUserOutputFields = []interface{}{
	"ID", "Number",
	map[string][]interface{}{
		"Type": {"Name", "CurrencyCode"},
	}}

type HandlerParams struct {
	service     *service.AccountService
	userService *userService.UserService
	logger      log15.Logger
}

func NewHandlerParams(
	service *service.AccountService,
	userService *userService.UserService,
	logger log15.Logger,
) *HandlerParams {
	return &HandlerParams{service, userService, logger.New("service", "HandlerParams")}
}

func (p *HandlerParams) forAdmin(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.Account{})
	params.AllowSelectFields(showListOutputFields)
	params.AllowPagination()
	p.addIncludes(params)
	p.addSortings(params)
	p.addFilters(params)
	return params
}

func (p *HandlerParams) forUser(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.Account{})
	params.AllowSelectFields(showListForUserOutputFields)
	p.addIncludes(params)
	params.Pagination.PageSize = 0
	params.Includes.AddIncludes("type")
	return params
}

func (p *HandlerParams) forAdminCsv(query string) *list_params.ListParams {
	params := p.forAdmin(query)
	params.Pagination.PageSize = 0
	params.Includes.AddIncludes("type")
	params.Includes.AddIncludes("user")
	return params
}

func (p *HandlerParams) addIncludes(params *list_params.ListParams) {
	params.AllowIncludes([]string{"type", "user"})
	params.AddCustomIncludes("user", p.service.UserIncludes)
}

func (p *HandlerParams) addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"id", "number", "isActive", "createdAt",
		"type.name", "user.username"})
	params.AddCustomSortings("user.username", p.userUsernameSorting)
	params.AddCustomSortings("type.name", p.typeNameSorting)
}

//TODO: make sorting by username
func (p *HandlerParams) userUsernameSorting(direction string,
	params *list_params.ListParams) (string, error) {
	return "", errors.New("Sorting by username is not implemented")
}

func (p *HandlerParams) typeNameSorting(direction string,
	params *list_params.ListParams) (string, error) {
	params.AddLeftJoin("account_types", "accounts.type_id = account_types.id")
	return fmt.Sprintf("account_types.name %s", direction), nil
}

func (p *HandlerParams) addFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{"numberContains", "typeId", "isActive",
		"createdAtFrom", "createdAtTo", "allowDeposits", "allowWithdrawals", "userId", "accountType.currencyCode", "isIwtInstructionsAvailable"})
	params.AddCustomFilter("numberContains", p.numberContainsFilter)
	params.AddCustomFilter("createdAtFrom", list_params.DateFromFilter("accounts.created_at"))
	params.AddCustomFilter("createdAtTo", list_params.DateToFilter("accounts.created_at"))
	params.AddCustomFilter("isActive", list_params.BoolFilter("accounts.is_active"))
	params.AddCustomFilter("allowDeposits", list_params.BoolFilter("accounts.allow_deposits"))
	params.AddCustomFilter("allowWithdrawals", list_params.BoolFilter("accounts.allow_withdrawals"))
	params.AddCustomFilter("accountType.currencyCode", p.accountTypeCurrencyCodeFilter)
	params.AddCustomFilter("isIwtInstructionsAvailable", p.iwtInstructionsAvailable)
}

func (p *HandlerParams) numberContainsFilter(inputValues []string,
	_ *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	var uids []string
	if len(inputValues[0]) > 2 {
		users, err := p.userService.GetByProfileData(inputValues[0])
		if err == nil {
			uids = make([]string, len(users))
			for i, user := range users {
				uids[i] = user.UID
			}
		} else {
			p.logger.Error("cannot get users", "err", err)
		}
	}

	condition := "accounts.number LIKE ?"
	if len(uids) > 0 {
		uidsAsString := `"` + strings.Join(uids, `","`) + `"`
		condition = condition + " OR accounts.user_id IN (" + uidsAsString + ")"
	}

	return condition, fmt.Sprintf("%%%s%%", inputValues[0])
}

func (p *HandlerParams) accountTypeCurrencyCodeFilter(
	inputValues []string, params *list_params.ListParams) (
	dbConditionPart string, dbValues interface{}) {
	params.AddLeftJoin("account_types", "accounts.type_id = account_types.id")
	pair := list_params.FieldOperatorPair{Field: "account_types.currency_code", Operator: list_params.OperatorEq}
	filter := list_params.FilterListParameter{FieldOperatorPair: pair, Values: inputValues}
	return params.GetConditionPartFromUsualFilter(&filter)
}

func (p *HandlerParams) iwtInstructionsAvailable(
	inputValues []string, params *list_params.ListParams) (
	dbConditionPart string, dbValues interface{}) {
	params.AddLeftJoin("account_types", "accounts.type_id = account_types.id")
	params.AddLeftJoin("iwt_bank_accounts", "account_types.currency_code = iwt_bank_accounts.currency_code")
	params.SetGroupBy("accounts.id")
	return "iwt_bank_accounts.is_iwt_enabled = ?", true
}
