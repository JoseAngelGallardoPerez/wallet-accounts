package scheduled_transaction

import (
	"errors"
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/user"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	pb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

func ExecuteScheduledTransactions(
	scheduledTransactions []*ScheduledTransaction,
	repo *Repository,
	requestCreator *request.Creator,
	db *gorm.DB,
	logger log15.Logger,
) {
	logger = logger.New("Task", "ExecuteScheduledTransactions")

	systemUser := user.GetSystemUser()
	successfullyExecutedTransfersCount := 0
	for _, transaction := range scheduledTransactions {
		tx := db.Begin()
		req, err := executeTransaction(tx, transaction, requestCreator, &systemUser)
		if err != nil {
			logger.Error("failed to execute transaction", "error", err)
			tx.Rollback()
			continue
		}

		txRepo := repo.WrapContext(tx)
		err = txRepo.Updates(&ScheduledTransaction{
			Id:        transaction.Id,
			RequestId: req.Id,
			Status:    StatusExecuted,
		})

		if err != nil {
			tx.Rollback()
			logger.Crit("failed to update scheduled transaction", err)
			continue
		}
		tx.Commit()
		successfullyExecutedTransfersCount++
		if successfullyExecutedTransfersCount > 0 {
			logger.Info(
				"successfully executed scheduled transfers",
				"count",
				successfullyExecutedTransfersCount,
			)
		}
	}
}

func executeTransaction(
	tx *gorm.DB,
	transaction *ScheduledTransaction,
	requestCreator *request.Creator,
	onBehalfUser *pb.User,
) (*requestModel.Request, error) {
	if transaction.Status != StatusPending {
		return nil, errors.New("invalid transaction status: " + string(transaction.Status))
	}
	description := transaction.Reason.Description()

	if transaction.Amount.LessThan(decimal.Zero) {
		params := &form.DA{
			Amount:                 transaction.Amount.Abs().String(),
			AccountId:              *transaction.AccountId,
			CreditToRevenueAccount: pointer.ToBool(true),
			Description:            description,
		}

		return requestCreator.CreateDARequest(params, onBehalfUser, tx)
	}

	creditAccountId := *transaction.AccountId
	if transaction.Account.InterestAccountId != nil {
		creditAccountId = *transaction.Account.InterestAccountId
		description += "\nfrom #: " + fmt.Sprintf("%s", transaction.Account.Number)
	}

	params := &form.CA{
		Amount:                  transaction.Amount.Abs().String(),
		Description:             description,
		AccountId:               creditAccountId,
		ApplyIwtFee:             pointer.ToBool(false),
		DebitFromRevenueAccount: pointer.ToBool(true),
	}

	return requestCreator.CreateCARequest(params, onBehalfUser, tx)
}
