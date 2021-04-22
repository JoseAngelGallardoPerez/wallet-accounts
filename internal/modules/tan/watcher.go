package tan

import (
	"github.com/Confialink/wallet-accounts/internal/srvdiscovery"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
	notificationspb "github.com/Confialink/wallet-notifications/rpc/proto/notifications"
)

type Watcher struct {
	started        bool
	repository     *Repository
	subscriberRepo *SubscriberRepository
	service        *Service
	settings       *settings.Service
	slots          chan int
	userService    *service.UserService
	logger         log15.Logger
}

type NotificationMethod string

const (
	NotificationMethodInternalMessage = "internal_message"
	NotificationMethodSms             = "sms"
)

func allowedNotificationMethods() map[NotificationMethod]bool {
	return map[NotificationMethod]bool{
		NotificationMethodInternalMessage: true,
	}
}

type Parameters struct {
	UserIds            []string
	Quantity           uint
	NotificationMethod NotificationMethod
}

func NewWatcher(repository *Repository, service *Service, settings *settings.Service, subscriberRepo *SubscriberRepository, userService *service.UserService, logger log15.Logger) *Watcher {
	slots := make(chan int, 5) // maximum 5 simultaneously message sending
	return &Watcher{
		repository:     repository,
		service:        service,
		settings:       settings,
		subscriberRepo: subscriberRepo,
		slots:          slots,
		userService:    userService,
		logger:         logger,
	}
}

// Start subscribes watcher on tan use event, it also checks all users for out of TANs every 30 minutes
func (w *Watcher) Start() {
	if w.started {
		w.logger.Error("tan watcher is already started")
		return
	}
	w.started = true
	w.logger.Info("Starting tan watcher.")
	w.service.OnSuccessfulUse(w.onUse) // execute w.onUse every time when tan is successfully used
	intv := time.Duration(30 * time.Minute)
	timer := time.NewTimer(intv)
	go func() {
		for {
			select {
			case <-timer.C:
				w.checkAllUsers()
				timer.Reset(intv)
			}
		}
	}()
	w.checkAllUsers() // check all on start
}

// execute every time when tan is successfully used
func (w *Watcher) onUse(userId, tan string) {
	w.GenerateNewBatchIfNeed(userId)
}

// check conditions, generate and send a new batch of tans to an user
func (w *Watcher) GenerateNewBatchIfNeed(userId string) {
	logger := w.logger.New("method", "GenerateNewBatchIfNeed")
	qty, err := w.settings.Int64(SettingTanGenerateTriggerQtyInt64)
	if nil != err {
		logger.Error("tan watcher failed to retrieve setting tan_generate_trigger_qty", "error", err)
		return
	}
	currentQty, err := w.repository.CountByUserId(userId)
	if nil != err {
		logger.Error("tan watcher failed to count number of remaining tans", "error", err, "userId", userId)
		return
	}
	if currentQty < uint(qty) {
		w.GenerateAndMessage(Parameters{UserIds: []string{userId}})
	}
}

// generate and send new batches of tans to users
func (w *Watcher) GenerateAndMessage(params Parameters) (err error) {
	// TODO: remove tans functionality
	return errors.New("GenerateAndMessage() method is deprecated")

	logger := w.logger.New("method", "GenerateAndMessage")
	var qty int64

	if params.Quantity > 0 {
		qty = int64(params.Quantity)
	} else {
		qty, err = w.settings.Int64(SettingTanGenerateQtyInt64)
		if nil != err {
			logger.Error("tan watcher failed to retrieve setting tan_generate_qty", "error", err)
			return
		}
	}

	if params.NotificationMethod == "" {
		params.NotificationMethod = NotificationMethodInternalMessage
	} else {
		if _, allowed := allowedNotificationMethods()[params.NotificationMethod]; !allowed {
			logger.Error("tan watcher failed to send notification: notification method is not allowed", "notificationMethod", params.NotificationMethod)
			err = errcodes.CreatePublicError(errcodes.CodeTanNotificationMethodNotAllowed)
			return
		}
	}

	notificationUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNameNotifications)
	if nil != err {
		logger.Error("tan watcher failed to discover notifications service", "error", err)
		return
	}
	notificationsClient := notificationspb.NewNotificationHandlerProtobufClient(notificationUrl.String(), http.DefaultClient)

	for _, uid := range params.UserIds {
		tans, err := w.service.GenerateAndAdd(uid, uint(qty))
		if nil != err {
			logger.Error("tan watcher failed to generate tans", "error", err, "uid", uid)
			return err
		}

		w.slots <- 1
		go func(uid string, tans []string) {
			w.sendNotifications(notificationsClient, uid, strings.Join(tans, "\n"), params.NotificationMethod)
			<-w.slots
		}(uid, tans)
	}
	return
}

func (w *Watcher) checkAllUsers() {
	logger := w.logger.New("method", "checkAllUsers")
	qty, err := w.settings.Int64(SettingTanGenerateTriggerQtyInt64)
	if nil != err {
		logger.Error("tan watcher failed to retrieve setting tan_generate_trigger_qty", "error", err)
		return
	}
	countResults, err := w.repository.FindUserIdsHavingTansLessThan(uint(qty))
	if nil != err {
		logger.Error("tan watcher failed to check all users", "error", err)
		return
	}
	zeroTansUids, err := w.subscriberRepo.UserIdsHavingZeroTans()
	if nil != err {
		logger.Error("tan watcher failed to check all users", "error", err)
		return
	}
	ids := make([]string, 0, len(countResults)+len(zeroTansUids))
	for _, countItem := range countResults {
		ids = append(ids, countItem.Uid)
	}
	ids = append(ids, zeroTansUids...)
	w.GenerateAndMessage(Parameters{UserIds: ids})
}

func (w *Watcher) sendNotifications(
	client notificationspb.NotificationHandler,
	userId,
	tan string,
	notificationMethod NotificationMethod,
) {
	logger := w.logger.New("method", "sendNotifications")
	user, err := w.userService.GetByUID(userId)
	if err != nil {
		logger.Error("failed to retrieve user", "error", err, "userId", userId)
		return
	}

	req := &notificationspb.Request{
		To:        user.UID,
		EventName: "TanCreate",
		Notifiers: []string{string(notificationMethod)},
		TemplateData: &notificationspb.TemplateData{
			Tan: tan,
		},
	}

	_, err = client.Dispatch(context.Background(), req)
	if nil != err {
		logger.Error("can't send notification", "error", err)
	}
}
