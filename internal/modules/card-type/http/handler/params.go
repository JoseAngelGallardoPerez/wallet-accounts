package handler

import (
	"fmt"

	"github.com/Confialink/wallet-pkg-list_params"

	"github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
)

type handlerParams struct {
}

var listOutputFields = []interface{}{"Id", "Name", "CurrencyCode", "IconId",
	map[string][]interface{}{
		"Category": {"Id", "Name"},
	}, map[string][]interface{}{
		"Format": {"Id", "Name", "Code"},
	}}

func (p *handlerParams) list(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.CardType{})
	params.AllowSelectFields(listOutputFields)
	params.AllowPagination()
	addIncludes(params)
	allowSortings(params)
	addFilters(params)
	return params
}

func allowSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"id", "name", "currencyCode", "category.name"})
	params.AddCustomSortings("category.name", categoryNameSorting)
}

func addFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{"nameContains"})
	params.AddCustomFilter("nameContains", list_params.ContainsFilter("name"))
}

func addIncludes(params *list_params.ListParams) {
	params.AllowIncludes([]string{"category", "format"})
}

func categoryNameSorting(direction string,
	params *list_params.ListParams) (string, error,
) {
	params.AddLeftJoin("card_type_categories AS card_type_categories", "card_types.card_type_category_id = card_type_categories.id")
	return fmt.Sprintf(`card_type_categories.name %s`, direction), nil
}
