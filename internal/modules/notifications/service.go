package notifications

import (
	"context"
	"errors"
	"net/http"

	notificationspb "github.com/Confialink/wallet-notifications/rpc/proto/notifications"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	moneyRequestEvent "github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/event"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	txModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-accounts/internal/srvdiscovery"
)

type Service struct {
	logger log15.Logger
}

func NewService(logger log15.Logger) *Service {
	return &Service{logger: logger.New("service", "notifications")}
}

func (s *Service) GetSetting(name string) (Setting, error) {
	settings, err := s.GetSettings(name)
	setting := Setting{}
	if err != nil {
		return setting, err
	}
	if len(settings) > 0 {
		setting = settings[0]
	}
	return setting, nil
}

// GetSettings retrieves notification settings by given names
func (s *Service) GetSettings(names ...string) ([]Setting, error) {
	logger := s.logger.New("method", "GetSettings")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return nil, err
	}

	req := &notificationspb.SettingsRequest{
		SettingNames: names,
	}
	response, err := client.GetSettings(context.Background(), req)
	if err != nil {
		return nil, err
	}

	settings := make([]Setting, len(response.Settings))
	for i, setting := range response.Settings {
		settings[i] = Setting{
			Name:  setting.Name,
			Value: setting.Value,
		}
	}

	return settings, nil
}

func (s *Service) TriggerIncomingTransaction(userId string, account *model.Account, methods []string) error {
	logger := s.logger.New("method", "TriggerIncomingTransaction")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return err
	}

	_, err = client.Dispatch(context.Background(), &notificationspb.Request{
		To:        userId,
		EventName: "IncomingTransaction",
		TemplateData: &notificationspb.TemplateData{
			AccountNumber: account.Number,
		},
		Notifiers: methods,
	})

	return err
}

func (s *Service) TriggerOutgoingTransaction(userId string, transaction *txModel.Transaction) error {
	logger := s.logger.New("method", "TriggerOutgoingTransaction")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return err
	}

	if transaction == nil || transaction.Id == nil {
		err = errors.New("transaction is nil or not persisted")
		logger.Error("failed to send outgoing transaction notification", "error", err)
		return err
	}

	_, err = client.Dispatch(context.Background(), &notificationspb.Request{
		To:        userId,
		EventName: "OutgoingTransaction",
		TemplateData: &notificationspb.TemplateData{
			TransactionId: *transaction.Id,
		},
	})

	return err
}

func (s *Service) TriggerNewTransferRequest(userId string) error {
	logger := s.logger.New("method", "TriggerNewTransferRequest")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return err
	}

	_, err = client.Dispatch(context.Background(), &notificationspb.Request{
		EventName: "NewTransferRequest",
		To:        userId,
	})

	return err
}

func (s *Service) TriggerNewTransferRequestByAdmin(user, admin *users.User) error {
	logger := s.logger.New("method", "TriggerNewTransferRequestByAdmin")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return err
	}

	_, err = client.Dispatch(context.Background(), &notificationspb.Request{
		EventName: "NewTransferRequestByAdmin",
		To:        admin.UID, // exclude that admin
		TemplateData: &notificationspb.TemplateData{
			UserName:  user.GetUsername(),
			FirstName: admin.GetFirstName(),
			LastName:  admin.GetLastName(),
		},
	})

	return err
}

func (s *Service) TriggerRequestExecuted(request *requestModel.Request) error {
	logger := s.logger.New("method", "TriggerRequestExecuted")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return err
	}

	_, err = client.Dispatch(context.Background(), &notificationspb.Request{
		EventName: "RequestExecuted",
		To:        *request.UserId,
		TemplateData: &notificationspb.TemplateData{
			RequestId: *request.Id,
		},
	})

	return err
}

func (s *Service) TriggerRequestCancelled(userID string, requestID uint64) error {
	logger := s.logger.New("method", "TriggerRequestCancelled")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return err
	}

	_, err = client.Dispatch(context.Background(), &notificationspb.Request{
		EventName: "RequestCancel",
		To:        userID,
		TemplateData: &notificationspb.TemplateData{
			RequestId: requestID,
		},
	})

	return err
}

func (s *Service) TriggerNewMoneyRequest(eventContext *moneyRequestEvent.Context) error {
	logger := s.logger.New("method", "TriggerIncomingTransaction")
	client, err := s.getClient()
	if err != nil {
		logger.Error("failed to get pb client", "error", err)
		return err
	}

	_, err = client.Dispatch(context.Background(), &notificationspb.Request{
		To:        eventContext.RecipientUID,
		EventName: "NewMoneyRequest",
		TemplateData: &notificationspb.TemplateData{
			EntityID:       eventContext.MoneyRequestId,
			OwnerFirstName: eventContext.SenderFirstName,
			OwnerLastName:  eventContext.SenderLastName,
		},
	})

	if err != nil {
		logger.Error("failed to notify user", "error", err)
	}

	return err
}

func (s *Service) getClient() (notificationspb.NotificationHandler, error) {
	notificationsUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNameNotifications)
	if nil != err {
		return nil, err
	}
	return notificationspb.NewNotificationHandlerProtobufClient(notificationsUrl.String(), http.DefaultClient), nil
}
