package system_logs

import (
	accountTypeModel "github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	bankDetailsModel "github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	cardTypeModel "github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
)

type SystemLogsService struct {
	transactionLogCreator    *TransactionLogCreator
	iwtBankDetailsLogCreator *IwtBankDetailsLogCreator
	accountLogCreator        *AccountLogCreator
	accountTypeLogCreator    *AccountTypeLogCreator
	cardLogCreator           *CardLogCreator
	cardTypeLogCreator       *CardTypeLogCreator
}

func NewSystemLogsService(
	transactionLogCreator *TransactionLogCreator,
	iwtBankDetailsLogCreator *IwtBankDetailsLogCreator,
	accountLogCreator *AccountLogCreator,
	accountTypeLogCreator *AccountTypeLogCreator,
	cardLogCreator *CardLogCreator,
	cardTypeLogCreator *CardTypeLogCreator,
) *SystemLogsService {
	return &SystemLogsService{
		transactionLogCreator:    transactionLogCreator,
		iwtBankDetailsLogCreator: iwtBankDetailsLogCreator,
		accountLogCreator:        accountLogCreator,
		accountTypeLogCreator:    accountTypeLogCreator,
		cardLogCreator:           cardLogCreator,
		cardTypeLogCreator:       cardTypeLogCreator,
	}
}

func (s *SystemLogsService) LogManualTransactionAsync(request *model.Request) {
	go s.transactionLogCreator.LogManualTransaction(request)
}

func (s *SystemLogsService) LogRevenueManualTransactionAsync(request *model.Request) {
	go s.transactionLogCreator.LogRevenueManualTransaction(request)
}

func (s *SystemLogsService) LogCreateIwtBankDetailsAsync(
	iwtBankDetails *bankDetailsModel.IwtBankAccountModel,
	userId string,
) {
	go s.iwtBankDetailsLogCreator.LogCreateIwtBankDetails(iwtBankDetails, userId)
}

func (s *SystemLogsService) LogDeleteIwtBankDetailsAsync(
	iwtBankDetails *bankDetailsModel.IwtBankAccountModel,
	userId string,
) {
	go s.iwtBankDetailsLogCreator.LogDeleteIwtBankDetails(iwtBankDetails, userId)
}

func (s *SystemLogsService) LogModifyIwtBankDetailsAsync(
	iwtBankDetails *bankDetailsModel.IwtBankAccountModel,
	data *bankDetailsModel.IwtBankAccountModel,
	userId string,
) {
	go s.iwtBankDetailsLogCreator.LogModifyIwtBankDetails(iwtBankDetails, data, userId)
}

func (s *SystemLogsService) LogCreateAccountAsync(
	account *accountModel.Account,
	userId string,
) {
	go s.accountLogCreator.LogCreateAccount(account, userId)
}

func (s *SystemLogsService) LogModifyAccountAsync(
	old *accountModel.Account,
	new *accountModel.Account,
	userId string,
) {
	go s.accountLogCreator.LogModifyAccount(old, new, userId)
}

func (s *SystemLogsService) LogCreateAccountTypeAsync(
	accountType *accountTypeModel.AccountType,
	userId string,
) {
	go s.accountTypeLogCreator.LogCreateAccountType(accountType, userId)
}

func (s *SystemLogsService) LogModifyAccountTypeAsync(
	oldRec *accountTypeModel.AccountType,
	newRec *accountTypeModel.AccountType,
	userId string,
) {
	go s.accountTypeLogCreator.LogModifyAccountType(oldRec, newRec, userId)
}

func (s *SystemLogsService) LogCreateCardAsync(
	card *cardModel.Card,
	userId string,
) {
	go s.cardLogCreator.LogCreateCard(card, userId)
}

func (s *SystemLogsService) LogModifyCardAsync(
	old *cardModel.Card,
	new *cardModel.Card,
	userId string,
) {
	go s.cardLogCreator.LogModifyCard(old, new, userId)
}

func (s *SystemLogsService) LogCreateCardTypeAsync(
	cardType *cardTypeModel.CardType,
	userId string,
) {
	go s.cardTypeLogCreator.LogCreateCardType(cardType, userId)
}

func (s *SystemLogsService) LogModifyCardTypeAsync(
	oldRec *cardTypeModel.CardType,
	newRec *cardTypeModel.CardType,
	userId string,
) {
	go s.cardTypeLogCreator.LogModifyCardType(oldRec, newRec, userId)
}
