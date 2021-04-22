package balance

import (
	"github.com/jinzhu/gorm"
	"time"
)

const dateLayout = "2006-01-02 15:04:05"

const (
	sqlGeneralTotal = `
		SELECT SUM(amount) as amount, currency_code from (
			SELECT a.available_amount as amount, t.currency_code  FROM accounts a
					INNER JOIN account_types t on a.type_id = t.id
					WHERE a.user_id = ?
					UNION ALL
					SELECT ABS(tx.amount) as amount, t.currency_code FROM transactions tx
					INNER JOIN accounts a ON tx.account_id = a.id
					INNER JOIN account_types t on t.id = a.type_id
					WHERE a.user_id = ? AND tx.status = 'pending'			
			) AS aux
		GROUP BY currency_code`

	sqlTotalDebitedPerPeriod = `
		SELECT SUM(ABS(tx.amount)) as amount, t.currency_code FROM transactions tx
				INNER JOIN accounts a ON tx.account_id = a.id
				INNER JOIN account_types t on t.id = a.type_id
				WHERE 
				a.user_id = ? 
				AND tx.status IN ('pending', 'executed')
				AND tx.created_at BETWEEN ? AND ?
				AND tx.amount < 0
		GROUP BY t.currency_code `
)

type dbGeneralTotalAggregator struct {
	db     *gorm.DB
	userId string
}

// Aggregate collects all user accounts available balances and absolute values of pending transactions
// related to the user accounts
func (d *dbGeneralTotalAggregator) Aggregate() (AggregationResult, error) {
	result := AggregationResult{}
	err := d.db.Raw(sqlGeneralTotal, d.userId, d.userId).Scan(&result).Error
	return result, err
}

type dbTotalDebitedPerPeriod struct {
	db       *gorm.DB
	userId   string
	dateFrom time.Time
	dateTo   time.Time
}

// Aggregate aggregates all outgoing user transactions for a certain period of time
func (d *dbTotalDebitedPerPeriod) Aggregate() (AggregationResult, error) {
	result := AggregationResult{}
	err := d.db.
		Raw(
			sqlTotalDebitedPerPeriod,
			d.userId,
			d.dateFrom.Format(dateLayout),
			d.dateTo.Format(dateLayout),
		).Scan(&result).Error
	return result, err
}
