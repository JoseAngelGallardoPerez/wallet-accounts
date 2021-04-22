package balance

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"time"
)

// Aggregator is used in order to collect list of operations applied to a given balance
type Aggregator interface {
	Aggregate() (AggregationResult, error)
}

// AggregationResult is a list of items related to a balance
type AggregationResult []AggregationItem

// AggregationItem is a single balance item representation e.g. transaction, current balance amount or anything
// else that could be presented as amount in currency
type AggregationItem struct {
	ItemAmount       decimal.Decimal `gorm:"column:amount"`
	ItemCurrencyCode string          `gorm:"column:currency_code"`
}

// AggregationItem returns item amount
func (a *AggregationItem) Amount() decimal.Decimal {
	return a.ItemAmount
}

// CurrencyCode determines which currency is related to amount
func (a *AggregationItem) CurrencyCode() string {
	return a.ItemCurrencyCode
}

// AggregationService helps to obtain and reduce balance aggregation results
type AggregationService struct {
	reducer Reducer
	factory AggregationFactory
}

// NewAggregationService creates new instance of AggregationService
func NewAggregationService(reducer Reducer, factory AggregationFactory) *AggregationService {
	return &AggregationService{reducer: reducer, factory: factory}
}

// GeneralTotalByUserId provides general total balance of all user accounts
func (a *AggregationService) GeneralTotalByUserId(userId, outCurrencyCode string) (AggregationItem, error) {
	aggregator, err := a.factory.GeneralTotalByUserId(userId)
	if err != nil {
		return AggregationItem{}, errors.Wrap(err, "failed to obtain aggregator")
	}
	return a.reduce(aggregator, outCurrencyCode)
}

// TotalDebitedByUserPerPeriod provides sum of all outgoing transactions from a user account by specific time period
func (a *AggregationService) TotalDebitedByUserPerPeriod(
	userId string,
	from,
	till time.Time,
	outCurrencyCode string,
) (AggregationItem, error) {
	aggregator, err := a.factory.TotalDebitedByUserIdPerPeriod(userId, from, till)
	if err != nil {
		return AggregationItem{}, errors.Wrap(err, "failed to obtain aggregator")
	}
	return a.reduce(aggregator, outCurrencyCode)
}

// WrapContext makes a copy of the service with new DB context
func (a AggregationService) WrapContext(db *gorm.DB) *AggregationService {
	a.factory = a.factory.WrapContext(db)
	return &a
}

// Reduce applies service reducer to passed results
func (a *AggregationService) Reduce(in AggregationResult, toCurrencyCode string) (out AggregationItem, err error) {
	item, err := a.reducer.Reduce(in, toCurrencyCode)
	if err != nil {
		return AggregationItem{}, errors.Wrapf(
			err,
			"failed to reduce aggregation result to currency %s",
			toCurrencyCode,
		)
	}
	return item, nil
}

// reduce uses applies given reducer on results returned from passed aggregator
func (a *AggregationService) reduce(aggregator Aggregator, outCurrencyCode string) (AggregationItem, error) {
	result, err := aggregator.Aggregate()
	if err != nil {
		return AggregationItem{}, errors.Wrap(err, "failed to obtain aggregation result")
	}
	return a.Reduce(result, outCurrencyCode)
}
