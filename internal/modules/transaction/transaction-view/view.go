package transaction_view

import (
	authService "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-pkg-utils/value"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/inconshreveable/log15"
)

type View interface {
	View(transactions []*model.Transaction, currentUser *users.User, includes ...string) (map[string]interface{}, error)
	SetTransactionSerializer(txSerializer TxSerializer) View
}

type DefaultView struct {
	txSerializer TxSerializer
	authService  authService.AuthServiceInterface
	logger       log15.Logger
}

func NewDefaultView(
	txSerializer TxSerializer,
	authService authService.AuthServiceInterface,
	logger log15.Logger,
) View {
	return &DefaultView{txSerializer: txSerializer, authService: authService, logger: logger}
}

func (d *DefaultView) View(transactions []*model.Transaction, currentUser *users.User, includes ...string) (map[string]interface{}, error) {

	result := map[string]interface{}{}

	txList := make([]interface{}, 0, len(transactions))
	for _, tx := range transactions {
		if !value.FromBool(tx.IsVisible) {
			continue
		}
		// here we append only those transactions which user is allowed to view
		if d.authService.CanDynamic(currentUser, authService.ActionRead, authService.ResourceTransaction, tx) {
			txSerialized, err := d.txSerializer(tx, includes...)
			if err != nil {
				d.logger.Error("unable to serialize transaction", "error", err)
				return nil, err
			}
			txList = append(txList, txSerialized)
		}
	}

	result["transactions"] = txList
	return result, nil
}

func (d DefaultView) SetTransactionSerializer(txSerializer TxSerializer) View {
	d.txSerializer = txSerializer
	return &d
}
