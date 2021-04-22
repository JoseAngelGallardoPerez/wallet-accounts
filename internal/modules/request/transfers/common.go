package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	txModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"time"
)

func makeCreditable(curr transfer.Currency, value *decimal.Decimal, additional *decimal.Decimal) transfer.Creditable {
	w1 := makeTransferWallet(curr, value)
	if additional == nil {
		return w1
	}
	creditable, err := transfer.JoinCreditable(w1, makeTransferWallet(curr, additional))
	if err != nil {
		panic(err)
	}
	return creditable
}

func makeDebitable(curr transfer.Currency, value *decimal.Decimal, additional *decimal.Decimal) transfer.Debitable {
	w1 := makeTransferWallet(curr, value)
	if additional == nil {
		return w1
	}
	debitable, err := transfer.JoinDebitable(w1, makeTransferWallet(curr, additional))
	if err != nil {
		panic(err)
	}
	return debitable
}

func makeTransferWallet(currency transfer.Currency, value *decimal.Decimal) *transfer.Wallet {
	linkedBalance := transfer.NewLinkedBalance(value)
	return transfer.NewWallet(linkedBalance, currency)
}

func currencies(currencyProvider transfer.CurrencyProvider, request *requestModel.Request) (base, reference transfer.Currency, err error) {
	base, err = currencyProvider.Get(*request.BaseCurrencyCode)
	if err != nil {
		err = errors.Wrapf(
			err,
			"error occurred while trying to get base currency %s data",
			*request.BaseCurrencyCode,
		)
		return
	}
	if *request.BaseCurrencyCode == *request.ReferenceCurrencyCode {
		reference = base
		return
	}
	reference, err = currencyProvider.Get(*request.ReferenceCurrencyCode)
	if err != nil {
		err = errors.Wrapf(
			err,
			"error occurred while trying to get reference currency %s data",
			*request.ReferenceCurrencyCode,
		)
	}
	return
}

func saveTransactions(db *gorm.DB, transactions []*txModel.Transaction, status string) error {
	for _, transaction := range transactions {
		transaction.Status = &status
		err := db.Save(transaction).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// syncAndUpdateTransactions is used in order to apply changes from details to existing transactions
func syncAndUpdateTransactions(db *gorm.DB, details types.Details, transactions []*txModel.Transaction, status string) error {
	now := time.Now()
	for _, tx := range transactions {
		detail, ok := details[constants.Purpose(*tx.Purpose)]
		if !ok {
			return errors.Wrapf(
				ErrModificationNotAllowed,
				`Transaction purpose "%s" is not found. It is assumed that changes will only affect existing transactions.`,
				*tx.Purpose,
			)
		}
		newTx := detail.Transaction

		tx.Amount = newTx.Amount
		tx.ShowAmount = newTx.ShowAmount
		tx.AvailableBalanceSnapshot = newTx.AvailableBalanceSnapshot
		tx.CurrentBalanceSnapshot = newTx.CurrentBalanceSnapshot
		tx.Status = pointer.ToString(status)
		err := db.
			Table("transactions").
			Where("id = ?", tx.Id).
			Updates(map[string]interface{}{
				"amount":                     newTx.Amount,
				"show_amount":                newTx.ShowAmount,
				"available_balance_snapshot": newTx.AvailableBalanceSnapshot,
				"current_balance_snapshot":   newTx.CurrentBalanceSnapshot,
				"status":                     status,
				"updated_at":                 now,
			}).Error

		if err != nil {
			return errors.Wrapf(err, "failed to update transaction #%d", *tx.Id)
		}
		newTx.Id = tx.Id
		newTx.Status = pointer.ToString(status)
		details[constants.Purpose(*tx.Purpose)].Transaction = newTx
	}
	return nil
}

func updateRequestStatus(db *gorm.DB, request *requestModel.Request, status string) error {
	if request.Status != nil && *request.Status == status {
		return nil
	}
	now := time.Now()
	request.Status = &status
	request.StatusChangedAt = &now
	return db.Exec(
		"UPDATE `requests` SET `status` = ?, `status_changed_at` = ? WHERE id = ?",
		status,
		now,
		request.Id,
	).Error
}

func updateRequestStatusAndAmount(db *gorm.DB, request *requestModel.Request, status string) error {
	if request.Status != nil && *request.Status == status {
		return nil
	}
	now := time.Now()
	request.Status = &status
	request.StatusChangedAt = &now
	return db.Exec(
		"UPDATE `requests` SET `status` = ?, `amount` = ?, `status_changed_at` = ? WHERE id = ?",
		status,
		request.Amount,
		now,
		request.Id,
	).Error
}

func updateRequestAmountAndRate(db *gorm.DB, request *requestModel.Request) error {
	return db.Exec(
		"UPDATE `requests` SET  `amount` = ?, `rate` = ? WHERE id = ?",
		request.Amount,
		request.Rate,
		request.Id,
	).Error
}

func updateRequestStatusAndCancellationReason(db *gorm.DB, request *requestModel.Request, status, reason string) error {
	if request.Status != nil && *request.Status == status {
		return nil
	}
	now := time.Now()
	request.Status = &status
	request.StatusChangedAt = &now
	return db.Exec(
		"UPDATE `requests` SET `status` = ?, `cancellation_reason` = ?, `status_changed_at` = ? WHERE id = ?",
		status,
		reason,
		now,
		request.Id,
	).Error
}

func updateTransactionsStatusByRequestId(db *gorm.DB, requestId uint64, status string) error {
	return db.Exec("UPDATE `transactions` SET `status` = ? WHERE `request_id` = ?", status, requestId).Error
}

func updateAccount(db *gorm.DB, accounts ...*model.Account) error {
	for _, acc := range accounts {
		err := db.
			Table(acc.TableName()).
			Where("id = ?", acc.ID).
			Updates(&model.AccountPrivate{
				AvailableAmount: acc.AvailableAmount,
				Balance:         acc.Balance,
			}).Error
		if err != nil {
			return errors.Wrapf(err, "failed to update account #%d (%s)", acc.ID, acc.Number)
		}
	}
	return nil
}

func updateCard(db *gorm.DB, card *cardModel.Card) error {
	err := db.Exec("UPDATE `cards` SET `balance` = ? WHERE `id` = ?", card.Balance, card.Id).Error
	if err != nil {
		return errors.Wrapf(err, "failed to update card #%d", *card.Id)
	}
	return nil
}

func updateRevenueAccount(db *gorm.DB, accounts ...*model.RevenueAccountModel) error {
	for _, acc := range accounts {
		err := db.
			Table(acc.TableName()).
			Where("id = ?", acc.ID).
			Updates(&model.RevenueAccountModel{
				RevenueAccountPublic: model.RevenueAccountPublic{
					Balance: acc.Balance,
				},
				RevenueAccountPrivate: model.RevenueAccountPrivate{
					AvailableAmount: acc.AvailableAmount,
				},
			}).Error
		if err != nil {
			return errors.Wrapf(err, "failed to update revenue account #%d (%s)", acc.ID, acc.CurrencyCode)
		}
	}
	return nil
}

func loadTransactions(db *gorm.DB, requestId uint64) ([]*txModel.Transaction, error) {
	var transactions []*txModel.Transaction
	err := db.
		Model(txModel.Transaction{}).
		Where("request_id = ?", requestId).
		Find(&transactions).
		Error
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load transactions, request id = %d", requestId)
	}
	return transactions, nil
}
