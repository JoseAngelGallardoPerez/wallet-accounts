package policy

import (
	"github.com/Confialink/wallet-accounts/internal/modules/balance/model"
	"github.com/Confialink/wallet-accounts/internal/modules/policy"
	"github.com/Confialink/wallet-users/rpc/proto/users"
)

func ProvideClientViewBalanceSnapshot() policy.Policy {
	return func(snapshot interface{}, user *users.User) bool {
		snap := snapshot.(*model.Snapshot)
		if snap.UserId == nil {
			return false
		}
		return *snap.UserId == user.UID
	}
}
