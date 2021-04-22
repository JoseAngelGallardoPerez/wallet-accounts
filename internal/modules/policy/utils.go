package policy

import "github.com/Confialink/wallet-users/rpc/proto/users"

// AnyOf true if any of a given policies is true
func AnyOf(policies ...Policy) Policy {
	return func(data interface{}, user *users.User) bool {
		for _, p := range policies {
			if p(data, user) {
				return true
			}
		}
		return false
	}
}
