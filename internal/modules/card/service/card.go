package service

import (
	"errors"
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/modules/app/validator"
	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/shopspring/decimal"

	modelSerializer "github.com/Confialink/wallet-accounts/internal/modules/app/serializer"
	cardTypeFormatModel "github.com/Confialink/wallet-accounts/internal/modules/card-type-format/model"
	cardTypeService "github.com/Confialink/wallet-accounts/internal/modules/card-type/service"
	"github.com/Confialink/wallet-accounts/internal/modules/card/model"
	"github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
)

const lastNumbersCount = 4

type CardService struct {
	repo              repository.CardRepositoryInterface
	modelSerializer   modelSerializer.ModelSerializerInterface
	cardTypeService   *cardTypeService.CardTypeService
	validator         validator.Interface
	systemLogsService *system_logs.SystemLogsService
}

func NewCardService(
	repo repository.CardRepositoryInterface,
	modelSerializer modelSerializer.ModelSerializerInterface,
	cardTypeService *cardTypeService.CardTypeService,
	validator validator.Interface,
	systemLogsService *system_logs.SystemLogsService,
) *CardService {
	return &CardService{
		repo,
		modelSerializer,
		cardTypeService,
		validator,
		systemLogsService,
	}
}

// Creates card from passed serialized card
func (s *CardService) Create(serializedCard *model.SerializedCard, currentUser *userpb.User) (res *model.Card, err error) {
	res = &model.Card{
		Number:          serializedCard.Number,
		Status:          serializedCard.Status,
		CardTypeId:      serializedCard.CardTypeId,
		UserId:          serializedCard.UserId,
		ExpirationYear:  serializedCard.ExpirationYear,
		ExpirationMonth: serializedCard.ExpirationMonth,
		Balance:         &decimal.Zero,
	}

	if err = s.validator.Struct(res); err != nil {
		return nil, err
	}

	res, err = s.repo.Create(res)
	if err != nil {
		return nil, err
	}

	includes := list_params.Includes{}
	includes.AddIncludes("CardType")
	includes.AddIncludes("CardType.Format")
	includes.AddIncludes("CardType.Category")

	res, err = s.repo.Get(*res.Id, &includes)
	if err != nil {
		return nil, err
	}

	s.systemLogsService.LogCreateCardAsync(res, currentUser.UID)
	return
}

func (s *CardService) Get(id uint32, includes *list_params.Includes) (res *model.Card, err error) {
	return s.repo.Get(id, includes)
}

func (s *CardService) GetList(listParams *list_params.ListParams) ([]*model.Card, error) {
	return s.repo.GetList(listParams)
}

func (s *CardService) GetListCount(listParams *list_params.ListParams) (uint64, error) {
	return s.repo.GetListCount(listParams)
}

// Updates model by fields and returns loaded model
func (s *CardService) UpdateFields(
	id uint32, fields map[string]interface{}, currentUser *userpb.User) (*model.Card, error) {

	includes := list_params.Includes{}
	includes.AddIncludes("CardType")
	includes.AddIncludes("CardType.Format")
	includes.AddIncludes("CardType.Category")

	old, err := s.repo.Get(id, &includes)
	if err != nil {
		return nil, err
	}

	card, err := s.repo.UpdateFields(id, fields)
	if err != nil {
		return nil, err
	}

	card, err = s.repo.Get(id, &includes)
	if err != nil {
		return nil, err
	}

	s.systemLogsService.LogModifyCardAsync(old, card, currentUser.UID)

	return card, nil
}

func (s *CardService) UserIncludes(records []interface{}) error {
	cards := make([]*model.Card, len(records))
	for i, v := range records {
		cards[i] = v.(*model.Card)
	}
	return s.repo.FillUsers(cards)
}

// BulkCreate creates list of cards
func (s *CardService) BulkCreate(cards []*model.Card) ([]*model.Card, error) {
	for _, card := range cards {
		cardExist, _ := s.repo.GetByNumber(*card.Number, nil)
		if nil != cardExist {
			return nil, errors.New(fmt.Sprintf("card with number %s already exists", *card.Number))
		}
	}
	return s.repo.BulkCreate(cards)
}

func (s *CardService) GetNumber(card *model.Card) string {
	cardNumberStr := *card.Number
	if *card.CardType.Format.Code == cardTypeFormatModel.CodeSixteenNumeric {
		if len(cardNumberStr) < lastNumbersCount {
			return fmt.Sprintf("**** **** **** %s", cardNumberStr)
		}
		return fmt.Sprintf("**** **** **** %s", cardNumberStr[len(cardNumberStr)-lastNumbersCount:])
	}
	return cardNumberStr
}
