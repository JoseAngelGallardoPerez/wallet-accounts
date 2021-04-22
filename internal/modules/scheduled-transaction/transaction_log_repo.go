package scheduled_transaction

import (
	"github.com/jinzhu/gorm"
)

type TransactionLogRepository struct {
	db *gorm.DB
}

func NewScheduledTransactionLogRepository(db *gorm.DB) *TransactionLogRepository {
	return &TransactionLogRepository{db: db}
}

func (s *TransactionLogRepository) Create(log *ScheduledTransactionLog) error {
	return s.db.Create(log).Error
}

func (s TransactionLogRepository) WrapContext(db *gorm.DB) *TransactionLogRepository {
	s.db = db
	return &s
}
