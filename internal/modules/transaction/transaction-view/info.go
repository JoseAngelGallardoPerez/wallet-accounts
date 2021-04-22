package transaction_view

import (
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	cardsRepository "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	userService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-utils/value"
	"github.com/shopspring/decimal"
)

type TxSerializer func(transaction *model.Transaction, includes ...string) (interface{}, error)

func UserTxSerializer(transaction *model.Transaction, includes ...string) (interface{}, error) {
	if !value.FromBool(transaction.IsVisible) {
		return nil, nil
	}
	logger := logger.New("where", "UserTxSerializer")

	requestedIncludes := struct {
		sender      bool
		recipient   bool
		requestData bool
		fees        bool

		loadRequestData bool
	}{}

	for _, include := range includes {
		switch include {
		case "sender", "Sender":
			requestedIncludes.sender = true
			requestedIncludes.loadRequestData = true
		case "recipient", "Recipient":
			requestedIncludes.recipient = true
			requestedIncludes.loadRequestData = true
		case "requestData", "RequestData":
			requestedIncludes.requestData = true
		case "fees", "Fees":
			requestedIncludes.fees = true
		}
	}

	amount := transaction.Amount
	if transaction.ShowAmount != nil {
		amount = transaction.ShowAmount
	}

	balanceSnapshot := transaction.AvailableBalanceSnapshot
	if transaction.ShowAvailableBalanceSnapshot != nil {
		balanceSnapshot = transaction.ShowAvailableBalanceSnapshot
	}

	request, _ := requestsRepository.FindById(value.FromUint64(transaction.RequestId))

	transactions, _ := transactionsRepository.GetByRequestId(value.FromUint64(transaction.RequestId))

	getFees := func() []*model.Transaction {
		fees := make([]*model.Transaction, 0, 1)
		for _, transaction := range transactions {
			if transaction.IsExchangeMarginFee() || transaction.IsDefaultTransferFee() {
				fees = append(fees, transaction)
			}
		}
		return fees
	}

	getOutgoingTransaction := func() *model.Transaction {
		for _, transaction := range transactions {
			if transaction.IsTargetOutgoing() {
				return transaction
			}
		}
		return nil
	}

	getIncomingTransaction := func() *model.Transaction {
		for _, transaction := range transactions {
			if transaction.IsIncoming() {
				return transaction
			}
		}
		return nil
	}

	showRequestId := request.IsVisible != nil && *request.IsVisible

	result := map[string]interface{}{
		"id":              transaction.Id,
		"status":          transaction.Status,
		"description":     transaction.Description,
		"requestId":       transaction.RequestId,
		"amount":          amount,
		"type":            transaction.Type,
		"purpose":         transaction.Purpose,
		"createdAt":       transaction.CreatedAt,
		"updatedAt":       transaction.UpdatedAt,
		"balanceSnapshot": balanceSnapshot,
		"requestSubject":  request.Subject,
		"requestRate":     request.Rate,
		"statusChangedAt": request.StatusChangedAt,
		"showRequestId":   showRequestId,
	}

	if requestedIncludes.requestData || requestedIncludes.loadRequestData {
		if requestedIncludes.requestData {
			requestData, err := requestDataPresenter.Present(request)
			if err != nil {
				logger.Warn("failed to present request data", "error", err, "requestId", *request.Id)
			}
			result["requestData"] = requestData
		}
		sourceAccountId, ok := request.SourceAccountId()
		if (requestedIncludes.sender || requestedIncludes.fees) && ok {
			sourceAccount, err := accountsRepository.FindByID(uint64(sourceAccountId))
			if err != nil {
				logger.Error("unable to fetch account by id", "error", err, "id", sourceAccountId)
				return nil, err
			}

			if requestedIncludes.sender {
				user, err := usersService.GetByUID(sourceAccount.UserId)
				if err != nil {
					logger.Error("unable to fetch user", "error", "uid", sourceAccount.UserId)
					return nil, err
				}

				sender := map[string]interface{}{
					"account": map[string]string{
						"number":       sourceAccount.Number,
						"typeName":     sourceAccount.Type.Name,
						"currencyCode": sourceAccount.Type.CurrencyCode,
					},
					"profile": map[string]string{
						"firstName":   user.FirstName,
						"lastName":    user.LastName,
						"companyName": user.CompanyName,
					},
				}

				if target := getOutgoingTransaction(); target != nil {
					sender["transaction"] = map[string]interface{}{
						"id":     target.Id,
						"amount": target.Amount.String(),
					}
				}

				result["sender"] = sender
			}

			if requestedIncludes.fees {
				var total decimal.Decimal

				feeList := make([]interface{}, 0, 1)
				for _, fee := range getFees() {
					total = total.Add(*fee.Amount)

					println(fee.Amount.String())
					println(total.String())

					item := map[string]string{
						"amount":      fee.Amount.String(),
						"status":      *fee.Status,
						"description": *fee.Description,
						"purpose":     *fee.Purpose,
					}

					feeList = append(feeList, item)
				}

				result["fees"] = map[string]interface{}{
					"total": total.String(),
					"list":  feeList,
				}
			}
		}

		if destinationAccountId, ok := request.DestinationAccountId(); requestedIncludes.recipient && ok {
			destAccount, err := accountsRepository.FindByID(uint64(destinationAccountId))
			if err != nil {
				logger.Error("unable to fetch account by id", "error", err, "id", destinationAccountId)
				return nil, err
			}

			user, err := usersService.GetByUID(destAccount.UserId)
			if err != nil {
				logger.Error("unable to fetch user", "error", "uid", destAccount.UserId)
				return nil, err
			}

			recipient := map[string]interface{}{
				"account": map[string]string{
					"number":       destAccount.Number,
					"typeName":     destAccount.Type.Name,
					"currencyCode": destAccount.Type.CurrencyCode,
				},
				"profile": map[string]string{
					"firstName":   user.FirstName,
					"lastName":    user.LastName,
					"companyName": user.CompanyName,
				},
			}

			if source := getIncomingTransaction(); source != nil {
				recipient["transaction"] = map[string]interface{}{
					"id":     source.Id,
					"amount": source.Amount.String(),
				}
			}

			result["recipient"] = recipient
		}
	}

	return result, nil
}

func ProvideDefaultInfoSerializer(
	userService *userService.UserService,
	accountRepository *accountRepository.AccountRepository,
	revenueAccountRepository *accountRepository.RevenueAccountRepository,
	cardsRepository cardsRepository.CardRepositoryInterface,
) TxSerializer {
	return func(tx *model.Transaction, includes ...string) (interface{}, error) {
		serializedTx, _ := UserTxSerializer(tx, includes...)
		info := map[string]interface{}{
			"transaction": serializedTx,
		}

		if tx.RevenueAccountId != nil {
			revenueAccount, err := revenueAccountRepository.FindByID(*tx.RevenueAccountId)
			if err != nil {
				return info, err
			}

			info["revenueAccount"] = map[string]interface{}{
				"revenueAccount": map[string]interface{}{
					"id":           revenueAccount.ID,
					"currencyCode": revenueAccount.CurrencyCode,
				},
			}
		}

		if tx.AccountId != nil {
			account, err := accountRepository.FindByID(*tx.AccountId)
			if err != nil {
				return info, err
			}

			user, err := userService.GetByUID(account.UserId)
			if err != nil {
				return info, err
			}

			info["account"] = map[string]interface{}{
				"account": map[string]interface{}{
					"number":       account.Number,
					"typeId":       account.TypeID,
					"description":  account.Description,
					"currencyCode": account.Type.CurrencyCode,
				},
				"user": map[string]string{
					"UID":       user.GetUID(),
					"userName":  user.GetUsername(),
					"firstName": user.GetFirstName(),
					"lastName":  user.GetLastName(),
				},
			}
		}

		if tx.CardId != nil {
			includes := list_params.NewIncludes("")
			includes.AddIncludes("CardType")
			card, err := cardsRepository.Get(*tx.CardId, includes)
			if err != nil {
				return info, err
			}

			info["card"] = map[string]interface{}{
				"number":       card.Number,
				"currencyCode": card.CardType.CurrencyCode,
			}
		}

		return info, nil
	}
}
