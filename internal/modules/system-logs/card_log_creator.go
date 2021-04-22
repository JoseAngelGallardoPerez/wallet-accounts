package system_logs

import (
	"encoding/json"
	"time"

	"github.com/Confialink/wallet-accounts/internal/recovery"

	"github.com/Confialink/wallet-accounts/internal/modules/card/model"
	usersService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/inconshreveable/log15"
)

type CardLogCreator struct {
	logsServiceWrap *logsServiceWrap
	usersService    *usersService.UserService
	logger          log15.Logger
	recoverer       func()
}

func NewCardLogCreator(
	usersService *usersService.UserService,
	logger log15.Logger,
) *CardLogCreator {
	logger = logger.New("service", "SystemLogsService")
	return &CardLogCreator{
		logsServiceWrap: newLogsServiceWrap(logger.New("serviceWrap", "logsServiceWrap")),
		usersService:    usersService,
		logger:          logger,
		recoverer:       recovery.DefaultRecoverer(),
	}
}

func (c *CardLogCreator) LogCreateCard(
	card *model.Card, userId string,
) {
	defer c.recoverer()
	owner, err := c.usersService.GetByUID(*card.UserId)

	if err != nil {
		c.logger.Error("Can't load user", err)
		return
	}

	card.User = &model.User{
		Username: &owner.Username,
	}

	data, err := json.Marshal(card)
	if err != nil {
		c.logger.Error("Can't marshall json", err)
		return
	}

	c.logsServiceWrap.createLog(
		SubjectCreateCard,
		userId,
		card.CreatedAt.Format(time.RFC3339),
		DataTitleCardDetails+": "+*card.Number,
		data,
	)
}

func (c *CardLogCreator) LogModifyCard(
	old *model.Card, new *model.Card, userId string,
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
		SubjectModifyCard,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleCardDetails+": "+*old.Number,
		data,
	)
}
