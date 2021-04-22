package service

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"net/url"

	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/inconshreveable/log15"
	pkgErrors "github.com/pkg/errors"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/validator"
	cardTypeCategoryRepo "github.com/Confialink/wallet-accounts/internal/modules/card-type-category/repository"
	cardTypeFormatRepo "github.com/Confialink/wallet-accounts/internal/modules/card-type-format/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/repository"
	cardsRepository "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	currenciesService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
)

type CardTypeService struct {
	repo                 repository.CardTypeRepositoryInterface
	cardTypeCategoryRepo *cardTypeCategoryRepo.CardTypeCategoryRepository
	cardTypeFormatRepo   *cardTypeFormatRepo.CardTypeFormatRepository
	currenciesService    currenciesService.CurrenciesServiceInterface
	cardsRepository      cardsRepository.CardRepositoryInterface
	logger               log15.Logger
	validator            validator.Interface
	systemLogService     *system_logs.SystemLogsService
}

func NewCardTypeService(
	repo repository.CardTypeRepositoryInterface,
	cardTypeCategoryRepo *cardTypeCategoryRepo.CardTypeCategoryRepository,
	cardTypeFormatRepo *cardTypeFormatRepo.CardTypeFormatRepository,
	currenciesService currenciesService.CurrenciesServiceInterface,
	cardsRepository cardsRepository.CardRepositoryInterface,
	logger log15.Logger,
	validator validator.Interface,
	systemLogService *system_logs.SystemLogsService,
) *CardTypeService {
	return &CardTypeService{
		repo:                 repo,
		cardTypeCategoryRepo: cardTypeCategoryRepo,
		cardTypeFormatRepo:   cardTypeFormatRepo,
		currenciesService:    currenciesService,
		cardsRepository:      cardsRepository,
		logger:               logger.New("Service", "CardTypeService"),
		validator:            validator,
		systemLogService:     systemLogService,
	}
}

// checkNameAndCurrencyCode checked account type by currency code and name
func (s *CardTypeService) checkNameAndCurrencyCode(name, currencyCode string) error {
	_, err := s.repo.FindByNameAndCurrencyCode(name, currencyCode)

	if err == nil {
		return errcodes.CreatePublicError(errcodes.CardTypeNameIsDuplicated, "Name and currency are already in use.")
	}

	if gorm.IsRecordNotFoundError(err) {
		return nil
	}

	return pkgErrors.Wrap(err, "failed to create new card type")
}

func (s *CardTypeService) Create(serialized *model.SerializedCardType, currentUser *userpb.User) (*model.CardType, error) {
	newModel := model.CardType{
		Id:                 serialized.Id,
		Name:               serialized.Name,
		CurrencyCode:       serialized.CurrencyCode,
		IconId:             serialized.IconId,
		CardTypeCategoryId: serialized.CardTypeCategoryId,
		CardTypeFormatId:   serialized.CardTypeFormatId,
	}

	if err := s.validator.Struct(&newModel); err != nil {
		return nil, err
	}

	cardTypeCategory, _ := s.cardTypeCategoryRepo.FindByID(*newModel.CardTypeCategoryId)
	if nil == cardTypeCategory {
		return nil, errcodes.CreatePublicError(errcodes.CodeCardTypeCategoryNotFound, fmt.Sprintf("card type category #%d not found", *newModel.CardTypeCategoryId))
	}

	cardTypeFormat, _ := s.cardTypeFormatRepo.FindByID(*newModel.CardTypeFormatId)
	if nil == cardTypeFormat {
		return nil, errcodes.CreatePublicError(errcodes.CodeCardTypeFormatNotFound, "card type format not found")
	}

	err := s.checkNameAndCurrencyCode(*newModel.Name, *newModel.CurrencyCode)

	if err != nil {
		return nil, err
	}

	res, err := s.repo.Create(&newModel)
	if err != nil {
		s.logger.Error("Failed to create new card type", "error", err)
		return nil, err
	}

	includes := list_params.Includes{}
	includes.AddIncludes("Category")
	includes.AddIncludes("Format")

	res, err = s.repo.Get(*res.Id, &includes)
	if err != nil {
		return nil, err
	}

	s.systemLogService.LogCreateCardTypeAsync(res, currentUser.UID)

	return res, nil
}

// Updates model by fields and returns loaded model
func (s *CardTypeService) UpdateFields(
	id uint32, fields map[string]interface{}, currentUser *userpb.User) (*model.CardType, error) {
	cardTypeCategory, _ := s.cardTypeCategoryRepo.FindByID(*fields["CardTypeCategoryId"].(*uint32))
	if nil == cardTypeCategory {
		return nil, errcodes.CreatePublicError(errcodes.CodeCardTypeCategoryNotFound, "card type category not found")
	}

	cardTypeFormat, _ := s.cardTypeFormatRepo.FindByID(*fields["CardTypeFormatId"].(*uint32))
	if nil == cardTypeFormat {
		return nil, errcodes.CreatePublicError(errcodes.CodeCardTypeFormatNotFound, "card type format not found")
	}

	includes := list_params.Includes{}
	includes.AddIncludes("Category")
	includes.AddIncludes("Format")

	old, err := s.repo.Get(id, &includes)
	if err != nil {
		return nil, err
	}

	err = s.checkNameAndCurrencyCode(*fields["Name"].(*string), *fields["CurrencyCode"].(*string))

	if (err != nil) && ((*fields["Name"].(*string) != *old.Name) || (*fields["CurrencyCode"].(*string) != *old.CurrencyCode)) {
		return nil, err
	}

	_, err = s.repo.UpdateFields(id, fields)
	if err != nil {
		return nil, err
	}

	card, err := s.repo.Get(id, &includes)
	if err != nil {
		return nil, err
	}

	s.systemLogService.LogModifyCardTypeAsync(old, card, currentUser.UID)

	return s.repo.Get(id, nil)
}

func (s *CardTypeService) Get(id uint32, includes *list_params.Includes) (*model.CardType, error) {
	return s.repo.Get(id, includes)
}

// Returns array of records with count by applying pagination
func (s *CardTypeService) GetListWithCount(
	params url.Values) (items []*model.CardType, count uint64, err error) {
	if items, err = s.repo.FindByParams(params); err != nil {
		return
	}
	count, err = s.repo.CountByParams(params)
	return
}

func (self *CardTypeService) GetList(params *list_params.ListParams) (
	[]*model.CardType, error,
) {
	return self.repo.GetList(params)
}

func (self *CardTypeService) GetCount(params *list_params.ListParams) (
	int64, error,
) {
	return self.repo.GetCount(params)
}

func (self *CardTypeService) Delete(id uint32) errors.TypedError {
	cards := self.cardsRepository.GetListByCardTypeId(id)
	if len(cards) > 0 {
		return errcodes.CreatePublicError(errcodes.CodeCardTypeAssociatedWithCards,
			"this card type is associated with cards and cannot be deleted")
	} else {
		self.repo.Delete(id)
		return nil
	}
}
