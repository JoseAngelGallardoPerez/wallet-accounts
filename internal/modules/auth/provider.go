package auth

import (
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/auth/service"
	"github.com/Confialink/wallet-accounts/internal/modules/permission"
	permissionPolicy "github.com/Confialink/wallet-accounts/internal/modules/permission/policy"
	"github.com/Confialink/wallet-accounts/internal/modules/policy"
	requestPolicy "github.com/Confialink/wallet-accounts/internal/modules/request/policy"
	transactionPolicy "github.com/Confialink/wallet-accounts/internal/modules/transaction/policy"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"

	"github.com/inconshreveable/log15"
	"github.com/kildevaeld/go-acl"
)

func Providers() []interface{} {
	return []interface{}{
		//AuthServiceInterface
		func(
			acl *acl.ACL,
			transactionRepository *repository.TransactionRepository,
			accountRepository *accountRepository.AccountRepository,
			logger log15.Logger,
		) service.AuthServiceInterface {
			clientViewTransferRequest := requestPolicy.ProvideClientViewRequestPolicy(
				transactionRepository,
				accountRepository,
				logger.New("policy", "ClientViewRequest"),
			)
			clientViewTransaction := transactionPolicy.ProvideClientViewTransaction(
				accountRepository,
				transactionRepository,
				logger.New("policy", "ClientViewTransaction"),
			)
			createModifyIwtBankAccount := permissionPolicy.ProvideCheckSpecificPermission(permission.CreateModifyIwtBankAccounts)
			revenueManager := permissionPolicy.ProvideCheckSpecificPermission(permission.ManageRevenue)
			revenueViewer := permissionPolicy.ProvideCheckSpecificPermission(permission.ViewRevenue)
			adminViewAccount := policy.AnyOf(
				permissionPolicy.ProvideCheckSpecificPermission(permission.ViewAccounts),
				permissionPolicy.ProvideCheckSpecificPermission(permission.InitiateExecuteUserTransfers),
				permissionPolicy.ProvideCheckSpecificPermission(permission.ViewUserReports),
			)
			adminViewCard := permissionPolicy.ProvideCheckSpecificPermission(permission.ViewCards)
			adminViewSettings := permissionPolicy.ProvideCheckSpecificPermission(permission.ViewSettings)
			policies := &service.RequiredPolicies{
				CreateModifyIwtBankAccount: createModifyIwtBankAccount,
				ViewTransferRequest:        clientViewTransferRequest,
				ViewTransaction:            clientViewTransaction,
				RevenueManager:             revenueManager,
				RevenueViewer:              revenueViewer,
				ViewAccount:                adminViewAccount,
				ViewCard:                   adminViewCard,
				ViewSettings:               adminViewSettings,
			}

			return service.NewService(acl, policies)
		},
	}
}
