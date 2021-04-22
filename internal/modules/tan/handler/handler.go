package handler

import (
	"net/http"

	"github.com/Confialink/wallet-accounts/internal/modules/notifications"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"

	userpb "github.com/Confialink/wallet-users/rpc/proto/users"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/tan"
)

type Controller struct {
	service             *tan.Service
	watcher             *tan.Watcher
	contextService      appHttpService.ContextInterface
	notificationService *notifications.Service
	logger              log15.Logger
}

func NewController(
	service *tan.Service,
	watcher *tan.Watcher,
	contextService appHttpService.ContextInterface,
	notificationService *notifications.Service,
	logger log15.Logger,
) *Controller {
	return &Controller{
		service:             service,
		watcher:             watcher,
		contextService:      contextService,
		notificationService: notificationService,
		logger:              logger,
	}
}

func (c *Controller) GetOwnCount(ctx *gin.Context) {
	currentUser := c.contextService.MustGetCurrentUser(ctx)

	count, err := c.service.Count(currentUser.UID)
	if nil != err {
		errors.AddErrors(ctx, &errors.PrivateError{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response.New().SetData(&tan.Count{UserId: currentUser.UID, Quantity: count}))
}

func (c *Controller) GetCount(ctx *gin.Context) {
	userId := ctx.Param("userId")
	count, err := c.service.Count(userId)
	if nil != err {
		errors.AddErrors(ctx, &errors.PrivateError{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response.New().SetData(&tan.Count{UserId: userId, Quantity: count}))
}

func (c *Controller) Create(ctx *gin.Context) {
	form := &tan.CreateForm{}
	if err := ctx.ShouldBind(form); err != nil {
		errors.AddErrors(ctx, &errors.PrivateError{Message: err.Error()})
		return
	}

	userId := ctx.Param("userId")
	if form.CancelOld {
		c.service.Cancel(userId)
	}
	err := c.watcher.GenerateAndMessage(tan.Parameters{
		UserIds:            []string{userId},
		Quantity:           form.Quantity,
		NotificationMethod: tan.NotificationMethod(form.NotificationMethod),
	})

	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusCreated)
}

// UserRequestOne is used in order to provide user with ability to request TAN on demand ("Request TAN" button)
func (c *Controller) UserRequestOne(ctx *gin.Context) {
	user := c.contextService.MustGetCurrentUser(ctx)
	if !c.userCanRequestTAN(user) {
		errcodes.AddError(ctx, errcodes.CodeTanRequestNotAllowed)
		return
	}

	err := c.watcher.GenerateAndMessage(tan.Parameters{
		UserIds:            []string{user.UID},
		Quantity:           1,
		NotificationMethod: tan.NotificationMethodSms,
	})

	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusCreated)
}

// UserCanRequestTan check is currently logged user is allowed to request TAN
func (c *Controller) UserCanRequestTan(ctx *gin.Context) {
	user := c.contextService.MustGetCurrentUser(ctx)

	answer := struct {
		IsAllowed bool `json:"isAllowed"`
	}{c.userCanRequestTAN(user)}

	ctx.JSON(http.StatusOK, response.New().SetData(answer))
}

func (c *Controller) userCanRequestTAN(user *userpb.User) bool {
	logger := c.logger.New("method", "userCanRequestTAN")

	setting, err := c.notificationService.GetSetting("tan_use_plivo")
	if err != nil {
		logger.Error("failed to retrieve notifications setting", "setting", "tan_use_plivo", "error", err)
		return false
	}

	return setting.Value == "true"
}
