package scheduled_transaction

import (
	"fmt"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Service struct {
	scheduledTransactionRepository    *Repository
	scheduledTransactionLogRepository *TransactionLogRepository
	db                                *gorm.DB
	logger                            log15.Logger
}

func NewService(
	scheduledTransactionRepository *Repository,
	scheduledTransactionLogRepository *TransactionLogRepository,
	db *gorm.DB,
	logger log15.Logger,
) *Service {
	return &Service{
		scheduledTransactionRepository:    scheduledTransactionRepository,
		scheduledTransactionLogRepository: scheduledTransactionLogRepository,
		db:                                db,
		logger:                            logger.New("service", "ScheduledTransactionService"),
	}
}

// ScheduleTransfer schedules transfer amount
// note that for outgoing transfer (e.g. fees) amount must be negative and vice versa for incoming transfer
// (interest generation)
func (s *Service) ScheduleTransfer(params *ScheduleParams, tx *gorm.DB) error {
	var err error
	if tx == nil {
		tx = s.db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
				return
			}
			tx.Commit()
		}()
	}

	err = s.validate(params, tx)
	if err != nil {
		return err
	}

	logger := s.logger.New("method", "ScheduleTransfer")
	txRepo := s.scheduledTransactionRepository.WrapContext(tx)
	logRepo := s.scheduledTransactionLogRepository.WrapContext(tx)

	transaction, err := s.findOrCreateScheduledTransaction(params, tx)
	if err != nil {
		logger.Error("failed to find or create scheduled transaction", "error", err, "params", s.paramsToString(params))
		return err
	}

	newAmount := transaction.Amount.Add(params.Amount)
	err = txRepo.Updates(&ScheduledTransaction{
		Id:     transaction.Id,
		Amount: newAmount,
	})
	if err != nil {
		return err
	}

	err = logRepo.Create(&ScheduledTransactionLog{
		ScheduledTransactionId: transaction.Id,
		Amount:                 params.Amount,
		CreatedAt:              &params.Now,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) findOrCreateScheduledTransaction(params *ScheduleParams, tx *gorm.DB) (*ScheduledTransaction, error) {
	transaction, err := s.scheduledTransactionRepository.FindNextPendingByAccountIdAndReason(params.Account.ID, params.Reason)

	if err == gorm.ErrRecordNotFound {
		scheduleDate := findNextDate(params.Period, params.PaymentDay)
		if params.Month != nil {
			y, _, d := scheduleDate.Date()
			scheduleDate = time.Date(y, *params.Month, d, 0, 0, 0, 0, scheduleDate.Location())
		}

		transaction = &ScheduledTransaction{
			Reason:        params.Reason,
			AccountId:     &params.Account.ID,
			Amount:        decimal.Zero,
			Status:        StatusPending,
			ScheduledDate: &scheduleDate,
		}

		repo := s.scheduledTransactionRepository
		if tx != nil {
			repo = repo.WrapContext(tx)
		}
		err = repo.Create(transaction)
		if err != nil {
			return nil, err
		}
	}

	return transaction, nil
}

func (s *Service) validate(params *ScheduleParams, tx *gorm.DB) error {
	validator, ok := validators[params.Reason]
	if !ok {
		panic("validator is not found " + params.Reason)
	}

	return validator(params, s.scheduledTransactionRepository.WrapContext(tx), s.logger)
}

func (s *Service) paramsToString(params *ScheduleParams) string {
	return fmt.Sprintf(
		"{AccountId: %d, Reason: %s, Period: %s, Amount: %s, PaymentDay: %d}",
		params.Account.ID,
		params.Reason,
		params.Period,
		params.Amount.String(),
		params.PaymentDay,
	)
}
