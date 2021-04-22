package system_logs

import (
	"time"

	"github.com/Confialink/wallet-accounts/internal/recovery"
	jsoniter "github.com/json-iterator/go"

	"github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	"github.com/inconshreveable/log15"
)

type AccountTypeLogCreator struct {
	logsServiceWrap *logsServiceWrap
	logger          log15.Logger
	recoverer       func()
}

func NewAccountTypeLogCreator(
	logger log15.Logger,
) *AccountTypeLogCreator {
	logger = logger.New("service", "SystemLogsService")
	return &AccountTypeLogCreator{
		logsServiceWrap: newLogsServiceWrap(logger.New("serviceWrap", "logsServiceWrap")),
		logger:          logger,
		recoverer:       recovery.DefaultRecoverer(),
	}
}

func (a *AccountTypeLogCreator) LogCreateAccountType(
	accountType *model.AccountType, userId string,
) {
	defer a.recoverer()

	data, err := jsoniter.Marshal(accountType)
	if err != nil {
		a.logger.Error("Can't marshall json", err)
		return
	}

	a.logsServiceWrap.createLog(
		SubjectCreateAccountTypes,
		userId,
		accountType.CreatedAt.Format(time.RFC3339),
		DataTitleAccountTypeDetails+": "+accountType.Name,
		data,
	)
}

func (a *AccountTypeLogCreator) LogModifyAccountType(
	old *model.AccountType, new *model.AccountType, userId string,
) {
	defer a.recoverer()

	json, err := jsoniter.Marshal(map[string]interface{}{
		"old": old,
		"new": new,
	})
	if err != nil {
		a.logger.Error("Can't marshall json", err)
		return
	}

	a.logsServiceWrap.createLog(
		SubjectModifyAccountTypes,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleAccountTypeDetails+": "+old.Name,
		json,
	)
}

func (a *AccountTypeLogCreator) getBoolText(bool *bool, trueStr string, falseStr string) string {
	if nil != bool && *bool != false {
		return trueStr
	}
	return falseStr
}
