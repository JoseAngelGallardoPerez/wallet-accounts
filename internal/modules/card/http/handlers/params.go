package handlers

import (
	"fmt"

	"github.com/Confialink/wallet-pkg-list_params"

	"github.com/Confialink/wallet-accounts/internal/modules/card/model"
	"github.com/Confialink/wallet-accounts/internal/modules/card/service"
)

type HandlerParams struct {
	service *service.CardService
}

var showListOutputFields = []interface{}{
	"Id",
	"Number",
	"Status",
	"Balance",
	"CardTypeId",
	"UserId",
	"ExpirationYear",
	"ExpirationMonth",
	"CreatedAt",
	map[string][]interface{}{
		"CardType": {
			"Id",
			"Name",
			"CurrencyCode",
			"IconId",
			map[string][]interface{}{
				"Category": {
					"Id",
					"Name",
				},
			},
			map[string][]interface{}{
				"Format": {
					"Id",
					"Name",
					"Code",
				},
			}},
		"User": {
			"Id",
			"Username",
			"Email",
			"FirstName",
			"LastName",
		},
	}}

func NewHandlerParams(service *service.CardService) *HandlerParams {
	return &HandlerParams{service}
}

func (p *HandlerParams) forAdminCsv(query string) *list_params.ListParams {
	params := p.getOwnListParams(query)
	params.Includes.AddIncludes("cardType")
	params.Includes.AddIncludes("user")
	params.Includes.AddIncludes("cardType.category")
	params.Includes.AddIncludes("cardType.format")
	return params
}

func (p *HandlerParams) getListParams(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.SerializedCard{})
	params.AllowSelectFields(showListOutputFields)
	p.addFilters(params)
	p.addIncludes(params)
	p.addSortings(params)
	params.AllowPagination()

	return params
}

func (p *HandlerParams) getOwnListParams(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.SerializedCard{})
	params.AllowSelectFields(showListOutputFields)
	p.addOwnFilters(params)
	p.addIncludes(params)
	p.addSortings(params)
	params.Pagination.PageSize = 0

	return params
}

func (p *HandlerParams) addFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{"userId", "numberContains", "cardTypeId",
		"cardType.currencyCode", "status", "createdAtFrom", "createdAtTo"})
	params.AddCustomFilter("numberContains", list_params.ContainsFilter("number"))
	params.AddCustomFilter("cardType.currencyCode", p.cardTypeCurrencyCodeFilter)
	params.AddCustomFilter("createdAtFrom", list_params.DateFromFilter("cards.created_at"))
	params.AddCustomFilter("createdAtTo", list_params.DateToFilter("cards.created_at"))
}

func (p *HandlerParams) addOwnFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{"numberContains", "cardTypeId",
		"cardType.currencyCode", "status", "createdAtFrom", "createdAtTo"})
	params.AddCustomFilter("numberContains", list_params.ContainsFilter("number"))
	params.AddCustomFilter("cardType.currencyCode", p.cardTypeCurrencyCodeFilter)
	params.AddCustomFilter("createdAtFrom", list_params.DateFromFilter("cards.created_at"))
	params.AddCustomFilter("createdAtTo", list_params.DateToFilter("cards.created_at"))
}

func (p *HandlerParams) addIncludes(params *list_params.ListParams) {
	params.AllowIncludes([]string{"cardType", "user", "cardType.category", "cardType.format"})
	params.AddCustomIncludes("user", p.service.UserIncludes)
}

func (p *HandlerParams) addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"id", "number", "status", "createdAt",
		"cardType.name", "expirationDate", "cardType.currencyCode",
		"user.username"})
	params.AddCustomSortings("cardType.name", p.cardTypeNameSorting)
	params.AddCustomSortings("expirationDate", p.expirationDateSorting)
	params.AddCustomSortings("cardType.currencyCode", p.cardTypeCurrencyCodeSorting)
	params.AddCustomSortings("user.username", p.userUsernameSorting)
}

func (p *HandlerParams) expirationDateSorting(direction string,
	_ *list_params.ListParams) (string, error) {
	return fmt.Sprintf("expiration_year %s,expiration_month %s", direction, direction), nil
}

func (p *HandlerParams) cardTypeNameSorting(direction string,
	params *list_params.ListParams) (string, error,
) {
	params.AddLeftJoin("card_types", "cards.card_type_id = card_types.id")
	return fmt.Sprintf("card_types.name %s", direction), nil
}

func (p *HandlerParams) cardTypeCurrencyCodeSorting(direction string,
	params *list_params.ListParams) (string, error,
) {
	params.AddLeftJoin("card_types", "cards.card_type_id = card_types.id")
	return fmt.Sprintf("card_types.currency_code %s", direction), nil
}

//TODO: make sorting by username
func (p *HandlerParams) userUsernameSorting(direction string,
	params *list_params.ListParams) (string, error) {
	return "", nil
}

func (p *HandlerParams) cardTypeCurrencyCodeFilter(
	inputValues []string, params *list_params.ListParams) (
	dbConditionPart string, dbValues interface{},
) {
	params.AddLeftJoin("card_types", "cards.card_type_id = card_types.id")
	pair := list_params.FieldOperatorPair{Field: "card_types.currency_code", Operator: list_params.OperatorEq}
	filter := list_params.FilterListParameter{FieldOperatorPair: pair, Values: inputValues}
	return params.GetConditionPartFromUsualFilter(&filter)
}
