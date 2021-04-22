package commands

import (
	"log"
	"net/url"
	"strconv"

	"github.com/Confialink/wallet-accounts/internal/modules/request"
	scheduledTransaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"go.uber.org/dig"
)

var executeScheduledTransaction command = command{
	Name:        "execute-scheduled-transaction",
	Usage:       "execute-scheduled-transaction?id={id}",
	Description: "Executes scheduled transfer request e.g. line of credit, interest generation etc.",
	Handler: func(c *dig.Container, args url.Values) {
		idStr := args.Get("id")
		if idStr == "" {
			log.Fatal("parameter \"id\" is required\n usage: execute-scheduled-transaction?id={id}")
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Fatal("invalid argument \"id\" is provided, it must be positive integer number.")
		}

		c.Invoke(func(
			db *gorm.DB,
			scheduledTransactionsRepository *scheduledTransaction.Repository,
			requestCreator *request.Creator,
			logger log15.Logger,
		) {
			txRepo := scheduledTransactionsRepository
			tx, err := txRepo.FindByID(id)
			if err != nil {
				log.Fatal("unable to retrieve scheduled transaction: ", err)
			}

			scheduledTransaction.ExecuteScheduledTransactions(
				[]*scheduledTransaction.ScheduledTransaction{tx},
				scheduledTransactionsRepository,
				requestCreator,
				db,
				logger.New("module", "command"),
			)
		})
	},
}
