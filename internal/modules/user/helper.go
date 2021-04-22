package user

import (
	"github.com/Confialink/wallet-users/rpc/proto/users"
)

var systemUser = users.User{
	UID:       "@system",
	GroupId:   0,
	FirstName: "@system",
	Email:     "system@system",
	RoleName:  "@system",
	LastName:  "@system",
	Username:  "@system",
}

func GetSystemUser() users.User {
	return systemUser
}

func IsSystemUser(u *users.User) bool {
	return u.UID == "@system"
}
