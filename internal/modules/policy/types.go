package policy

import "github.com/Confialink/wallet-users/rpc/proto/users"

type Policy func(interface{}, *users.User) bool
