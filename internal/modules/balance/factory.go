package balance

import (
	"github.com/jinzhu/gorm"
	"time"
)

// AggregationFactory provides number of the aggregators
type AggregationFactory interface {
	// GeneralTotalByUserId is an aggregator which summarize all user account balances
	// including pending transactions (as absolute values)
	GeneralTotalByUserId(userId string) (Aggregator, error)
	// TotalDebitedByUserIdPerPeriod is an aggregator which summarize all user outgoing transaction
	// for particular period
	TotalDebitedByUserIdPerPeriod(userId string, from, till time.Time) (Aggregator, error)
	// WrapContext creates a copy of the factory with provided db context
	WrapContext(db *gorm.DB) AggregationFactory
}

type dbAggregationFactory struct {
	db *gorm.DB
}

func NewDBAggregationFactory(db *gorm.DB) AggregationFactory {
	return &dbAggregationFactory{db: db}
}

// GeneralTotalByUserId is an aggregator which summarize all user account balances
// including pending transactions (as absolute values)
func (d *dbAggregationFactory) GeneralTotalByUserId(userId string) (Aggregator, error) {
	return &dbGeneralTotalAggregator{
		db:     d.db,
		userId: userId,
	}, nil
}

// TotalDebitedByUserIdPerPeriod is an aggregator which summarize all user outgoing transaction
// for particular period
func (d *dbAggregationFactory) TotalDebitedByUserIdPerPeriod(userId string, from, till time.Time) (Aggregator, error) {
	return &dbTotalDebitedPerPeriod{
		db:       d.db,
		userId:   userId,
		dateFrom: from,
		dateTo:   till,
	}, nil
}

// WrapContext creates a copy of the factory
func (d dbAggregationFactory) WrapContext(db *gorm.DB) AggregationFactory {
	d.db = db
	return &d
}
