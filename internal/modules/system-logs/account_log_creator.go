package system_logs

import (
	"encoding/json"
	"time"

	userModel "github.com/Confialink/wallet-accounts/internal/modules/user/model"
	"github.com/Confialink/wallet-accounts/internal/recovery"

	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	usersService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/inconshreveable/log15"
)

type AccountLogCreator struct {
	logsServiceWrap *logsServiceWrap
	usersService    *usersService.UserService
	logger          log15.Logger
	recoverer       func()
}

func NewAccountLogCreator(
	usersService *usersService.UserService,
	logger log15.Logger,
) *AccountLogCreator {
	logger = logger.New("service", "SystemLogsService")
	return &AccountLogCreator{
		logsServiceWrap: newLogsServiceWrap(logger.New("serviceWrap", "logsServiceWrap")),
		usersService:    usersService,
		logger:          logger,
		recoverer:       recovery.DefaultRecoverer(),
	}
}

func (a *AccountLogCreator) LogCreateAccount(
	account *model.Account, userId string,
) {
	defer a.recoverer()
	owner, err := a.usersService.GetByUID(account.UserId)

	if err != nil {
		a.logger.Error("Can't load user", err)
		return
	}

	account.User = &userModel.User{
		Username: &owner.Username,
	}

	data, err := json.Marshal(account)
	if err != nil {
		a.logger.Error("Can't marshall json", err)
		return
	}

	a.logsServiceWrap.createLog(
		SubjectCreateAccount,
		userId,
		account.CreatedAt.Format(time.RFC3339),
		DataTitleAccountDetails+": "+account.Number,
		data,
	)
}

func (a *AccountLogCreator) LogModifyAccount(
	old *model.Account, new *model.Account, userId string,
) {
	defer a.recoverer()

	data, err := json.Marshal(map[string]interface{}{
		"old": old,
		"new": new,
	})
	if err != nil {
		a.logger.Error("Can't marshall json", err)
		return
	}

	a.logsServiceWrap.createLog(
		SubjectModifyAccount,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleAccountDetails+": "+old.Number,
		data,
	)
}
