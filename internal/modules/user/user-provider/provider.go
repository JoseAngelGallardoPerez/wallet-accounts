package user_provider

import "github.com/Confialink/wallet-accounts/internal/modules/user/service"

func Providers() []interface{} {
	return []interface{}{
		//user.service.UserService
		service.NewUserService,
	}
}
