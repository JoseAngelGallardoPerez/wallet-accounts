package handler

import (
	"github.com/Confialink/wallet-accounts/internal/modules/notifications"
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/inconshreveable/log15"
)

var (
	logger              log15.Logger
	notificationService *notifications.Service
	usersService        *service.UserService
	settingsService     *settings.Service
)

func LoadDependencies(
	loggerDep log15.Logger,
	notificationServiceDep *notifications.Service,
	usersServiceDep *service.UserService,
	settingsServiceDep *settings.Service,
) {
	logger = loggerDep
	notificationService = notificationServiceDep
	usersService = usersServiceDep
	settingsService = settingsServiceDep
}
