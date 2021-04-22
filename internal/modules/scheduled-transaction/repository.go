package scheduled_transaction

import (
	"time"

	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/jinzhu/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) Create(transaction *ScheduledTransaction) error {
	return s.db.Create(transaction).Error
}

func (s *Repository) Updates(transaction *ScheduledTransaction) error {
	return s.db.Model(transaction).Updates(transaction).Error
}

func (s *Repository) GetScheduled(reason string, now time.Time) ([]*ScheduledTransaction, error) {
	var result []*ScheduledTransaction

	err := s.db.
		Preload("Account").
		Preload("Account.Type").
		Model(&ScheduledTransaction{}).
		Find(
			&result,
			"status = ? and reason = ? and scheduled_date < ?",
			StatusPending,
			reason,
			now,
		).Error

	return result, err
}

// FindNextPendingByAccountIdAndReason looking scheduled for future transaction with the given account id and reason
// having status "pending"
func (s *Repository) FindNextPendingByAccountIdAndReason(
	accountId uint64,
	reason Reason,
) (*ScheduledTransaction, error) {

	scheduledTransaction := &ScheduledTransaction{}
	err := s.db.Limit(1).Find(
		scheduledTransaction,
		"account_id = ? AND reason = ? AND status = ? AND scheduled_date > ?",
		accountId,
		reason,
		StatusPending,
		time.Now(),
	).Error

	return scheduledTransaction, err
}

func (s Repository) WrapContext(db *gorm.DB) *Repository {
	s.db = db
	return &s
}

// GetList returns records from passed ListParams
func (s *Repository) GetList(params *list_params.ListParams) (
	[]*ScheduledTransaction, error,
) {
	var items []*ScheduledTransaction

	str, arguments := params.GetWhereCondition()
	query := s.db.Where(str, arguments...)

	query = query.Order(params.GetOrderByString())

	if params.GetLimit() != 0 {
		query = query.Limit(params.GetLimit())
	}
	query = query.Offset(params.GetOffset())

	query = query.Joins(params.GetJoinCondition())

	for _, preloadName := range params.GetPreloads() {
		query = query.Preload(preloadName)
	}

	if err := query.Find(&items).Error; err != nil {
		return items, err
	}

	interfaceScheduledTransactions := make([]interface{}, len(items))
	for i, itemPtr := range items {
		interfaceScheduledTransactions[i] = itemPtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceScheduledTransactions); err != nil {
			return items, err
		}
	}

	return items, nil
}

// GetListCount returns count of scheduled transactions
// receive *list_params.ListParams
func (s *Repository) GetListCount(params *list_params.ListParams,
) (uint64, error) {
	var count uint64

	query := s.db.Joins(params.GetJoinCondition())
	str, arguments := params.GetWhereCondition()
	query = query.Where(str, arguments...)

	if err := query.Model(&ScheduledTransaction{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

// FindByID find scheduled transaction by id
func (repo *Repository) FindByID(id uint64) (*ScheduledTransaction, error) {
	var transaction ScheduledTransaction
	transaction.Id = &id
	if err := repo.db.
		Preload("Account").
		Preload("Account.Type").
		First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}
