package policy

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
)

// CanClientRead checks if client can read an account
func CanClientRead(account interface{}, user *userpb.User) bool {
	a := account.(*model.Account)
	return nil != a && a.UserId == user.UID
}

// CanClientReadList checks if client can read list of accounts
func CanClientReadList(account interface{}, user *userpb.User) bool {
	return true
}
