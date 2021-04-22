package protoserv

import (
	"github.com/Confialink/wallet-pkg-list_params"
	"context"
	"errors"
	"github.com/inconshreveable/log15"

	accountTypeModel "github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	accountTypesRepository "github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	accountSrv "github.com/Confialink/wallet-accounts/internal/modules/account/service"
	bankDetailModel "github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	bankDetailsRepository "github.com/Confialink/wallet-accounts/internal/modules/bank-details/repository"
	cardTypeModel "github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	cardTypesRepository "github.com/Confialink/wallet-accounts/internal/modules/card-type/repository"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	cardRepository "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	settingsRepository "github.com/Confialink/wallet-accounts/internal/modules/settings/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/tan"
	pb "github.com/Confialink/wallet-accounts/rpc/accounts"
)

type ProtobufServer struct {
	cardTypesRepository    cardTypesRepository.CardTypeRepositoryInterface
	accountTypesRepository *accountTypesRepository.AccountTypeRepository
	bankDetailsRepository  *bankDetailsRepository.IwtBankAccountRepository
	settingsRepository     *settingsRepository.Settings
	cardRepository         cardRepository.CardRepositoryInterface
	accountRepository      *accountRepository.AccountRepository
	accountService         *accountSrv.AccountService
	tanWatcher             *tan.Watcher
	logger                 log15.Logger
}

func NewProtobufServer(
	cardTypesRepository cardTypesRepository.CardTypeRepositoryInterface,
	accountTypesRepository *accountTypesRepository.AccountTypeRepository,
	bankDetailsRepository *bankDetailsRepository.IwtBankAccountRepository,
	settingsRepository *settingsRepository.Settings,
	cardRepository cardRepository.CardRepositoryInterface,
	accountRepository *accountRepository.AccountRepository,
	accountService *accountSrv.AccountService,
	tanWatcher *tan.Watcher,
	logger log15.Logger,
) *ProtobufServer {
	return &ProtobufServer{
		cardTypesRepository,
		accountTypesRepository,
		bankDetailsRepository,
		settingsRepository,
		cardRepository,
		accountRepository,
		accountService,
		tanWatcher,
		logger,
	}
}

// CanDisableCurrency checks if currency can be disabled by code
func (server *ProtobufServer) CanDisableCurrency(ctx context.Context,
	req *pb.DisableCurrencyReq,
) (response *pb.DisableCurrencyResp, err error) {
	response = &pb.DisableCurrencyResp{}

	canCardTypes, err := server.canDisableForCardTypes(req.Code)
	if err != nil {
		return
	}

	canAccountTypes, err := server.canDisableForAccountTypes(req.Code)
	if err != nil {
		return
	}

	canIWTBankAccounts, err := server.canDisableForIWTBankDetails(req.Code)
	if err != nil {
		return
	}

	response.Can = canCardTypes && canAccountTypes && canIWTBankAccounts
	return
}

func (server *ProtobufServer) GetSettingsByName(ctx context.Context,
	req *pb.SettingsByNameReq,
) (response *pb.SettingsByNameResp, err error) {
	response = &pb.SettingsByNameResp{}

	value, err := server.settingsRepository.FirstByName(req.Name)
	if err != nil {
		return
	}

	response.Value = value.Value
	return
}

func (server *ProtobufServer) GenerateAndSendTans(ctx context.Context,
	req *pb.GenerateAndSendTansReq,
) (*pb.GenerateAndSendTansResp, error) {

	return &pb.GenerateAndSendTansResp{}, errors.New("deprecated method")
}

func (server *ProtobufServer) UserHasCardsOrAccountsBy(ctx context.Context, req *pb.UserHasCardsOrAccountsReq) (resp *pb.UserHasCardsOrAccountsResp, err error) {
	resp = &pb.UserHasCardsOrAccountsResp{}
	params := list_params.NewListParamsFromQuery("", cardModel.Card{})
	params.AddFilter("user_id", []string{req.Uid})
	cardsCount, err := server.cardRepository.GetListCount(params)
	if err != nil {
		server.logger.Error("Failed to get count of cards", "error", err)
		return
	}

	if cardsCount > 0 {
		resp.CardsExist = true
	}

	params = list_params.NewListParamsFromQuery("", accountModel.Account{})
	params.AddFilter("user_id", []string{req.Uid})
	accountsCount, err := server.accountRepository.GetCount(params)
	if err != nil {
		server.logger.Error("Failed to get count of accounts", "error", err)
		return
	}

	if accountsCount > 0 {
		resp.AccountsExist = true
	}

	return
}

func (server *ProtobufServer) GenerateAccount(ctx context.Context, req *pb.GenerateAccountReq) (resp *pb.GenerateAccountResp, err error) {
	resp = &pb.GenerateAccountResp{}

	res, err := server.accountService.CreateAccountAsWallet(req.CurrencyCode, req.Uid)
	if err != nil {
		return
	}

	resp.Id = res.ID
	resp.Number = res.Number
	return
}

func (server *ProtobufServer) canDisableForCardTypes(code string) (bool, error) {
	params := list_params.NewListParamsFromQuery("", cardTypeModel.CardType{})
	params.AddFilter("currencyCode", []string{code})

	cardTypesCount, err := server.cardTypesRepository.GetCount(params)
	if err != nil {
		server.logger.Error("Failed to get card types count", "error", err)
		return false, errors.New("failed to get card types count")
	}

	return cardTypesCount == 0, nil
}

func (server *ProtobufServer) canDisableForAccountTypes(code string) (bool, error) {
	params := list_params.NewListParamsFromQuery("", accountTypeModel.AccountType{})
	params.AddFilter("currencyCode", []string{code})

	accountTypesCount, err := server.accountTypesRepository.GetListCount(params)
	if err != nil {
		server.logger.Error("Failed to get account types count", "error", err)
		return false, errors.New("failed to get account types count")
	}

	return accountTypesCount == 0, nil
}

func (server *ProtobufServer) canDisableForIWTBankDetails(code string) (bool, error) {
	params := list_params.NewListParamsFromQuery("", bankDetailModel.IwtBankAccountModel{})
	params.AddFilter("currencyCode", []string{code})

	accountTypesCount, err := server.bankDetailsRepository.GetListCount(params)
	if err != nil {
		server.logger.Error("Failed to get iwt bank accounts count", "error", err)
		return false, errors.New("failed to get iwt bank accounts count")
	}

	return accountTypesCount == 0, nil
}
