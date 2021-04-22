package system_logs

import (
	"encoding/json"
	"time"

	"github.com/Confialink/wallet-accounts/internal/recovery"

	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	"github.com/inconshreveable/log15"
)

type IwtBankDetailsLogCreator struct {
	logsServiceWrap *logsServiceWrap
	logger          log15.Logger
	recoverer       func()
}

func NewIwtBankDetailsLogCreator(
	logger log15.Logger,
) *IwtBankDetailsLogCreator {
	logger = logger.New("service", "SystemLogsService")
	return &IwtBankDetailsLogCreator{
		logsServiceWrap: newLogsServiceWrap(logger.New("serviceWrap", "logsServiceWrap")),
		logger:          logger,
		recoverer:       recovery.DefaultRecoverer(),
	}
}

func (l *IwtBankDetailsLogCreator) LogCreateIwtBankDetails(
	iwtBankDetails *model.IwtBankAccountModel, userId string,
) {
	defer l.recoverer()

	data, err := json.Marshal(iwtBankDetails)
	if err != nil {
		l.logger.Error("Can't marshall json", err)
		return
	}

	l.logsServiceWrap.createLog(
		SubjectCreateIwtBankAccounts,
		userId,
		iwtBankDetails.CreatedAt.Format(time.RFC3339),
		DataTitleBankAccountReferenceDetails+": "+iwtBankDetails.BeneficiaryBankDetails.BankName,
		data,
	)
}

func (l *IwtBankDetailsLogCreator) LogDeleteIwtBankDetails(
	iwtBankDetails *model.IwtBankAccountModel, userId string,
) {
	defer l.recoverer()

	data, err := json.Marshal(iwtBankDetails)
	if err != nil {
		l.logger.Error("Can't marshall json", err)
		return
	}

	l.logsServiceWrap.createLog(
		SubjectDeleteIwtBankAccounts,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleBankAccountReference+": "+iwtBankDetails.BeneficiaryBankDetails.BankName,
		data,
	)
}

func (l *IwtBankDetailsLogCreator) LogModifyIwtBankDetails(
	old *model.IwtBankAccountModel,
	new *model.IwtBankAccountModel,
	userId string,
) {
	defer l.recoverer()

	data, err := json.Marshal(map[string]interface{}{
		"old": old,
		"new": new,
	})

	if err != nil {
		l.logger.Error("Can't marshall json", err)
		return
	}

	l.logsServiceWrap.createLog(
		SubjectModifyIwtBankAccounts,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleBankAccountReferenceDetails+": "+old.BeneficiaryBankDetails.BankName,
		data,
	)
}
