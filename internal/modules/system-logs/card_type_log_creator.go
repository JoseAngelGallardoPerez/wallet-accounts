package system_logs

import (
	"encoding/json"
	"time"

	"github.com/Confialink/wallet-accounts/internal/recovery"

	"github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	"github.com/inconshreveable/log15"
)

type CardTypeLogCreator struct {
	logsServiceWrap *logsServiceWrap
	logger          log15.Logger
	recoverer       func()
}

func NewCardTypeLogCreator(
	logger log15.Logger,
) *CardTypeLogCreator {
	logger = logger.New("service", "SystemLogsService")
	return &CardTypeLogCreator{
		logsServiceWrap: newLogsServiceWrap(logger.New("serviceWrap", "logsServiceWrap")),
		logger:          logger,
		recoverer:       recovery.DefaultRecoverer(),
	}
}

func (c *CardTypeLogCreator) LogCreateCardType(
	cardType *model.CardType, userId string,
) {
	defer c.recoverer()

	data, err := json.Marshal(cardType)
	if err != nil {
		c.logger.Error("Can't marshall json", err)
		return
	}

	c.logsServiceWrap.createLog(
		SubjectCreateCardType,
		userId,
		cardType.CreatedAt.Format(time.RFC3339),
		DataTitleCardTypeDetails+": "+*cardType.Name,
		data,
	)
}

func (c *CardTypeLogCreator) LogModifyCardType(
	old *model.CardType, new *model.CardType, userId string,
) {
	defer c.recoverer()

	data, err := json.Marshal(map[string]interface{}{
		"old": old,
		"new": new,
	})
	if err != nil {
		c.logger.Error("Can't marshall json", err)
		return
	}

	c.logsServiceWrap.createLog(
		SubjectModifyCardType,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleCardTypeDetails+": "+*old.Name,
		data,
	)
}
