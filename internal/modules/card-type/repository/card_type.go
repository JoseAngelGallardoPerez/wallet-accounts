package repository

import (
	"net/url"
	"strconv"

	"github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	currencyModel "github.com/Confialink/wallet-accounts/internal/modules/currency/model"
	currenciesService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/iancoleman/strcase"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

const DefaultLimit = 15
const DefaultOffset = 0

type CardTypeRepositoryInterface interface {
	Create(*model.CardType) (*model.CardType, error)
	UpdateFields(uint32, map[string]interface{}) (*model.CardType, error)
	FindByNameAndCurrencyCode(name, currencyCode string) (*model.CardType, error)
	Get(uint32, *list_params.Includes) (*model.CardType, error)
	FindByParams(url.Values) ([]*model.CardType, error)
	CountByParams(url.Values) (uint64, error)
	GetList(*list_params.ListParams) ([]*model.CardType, error)
	GetCount(*list_params.ListParams) (int64, error)
	Delete(id uint32)
}

type cardTypeRepository struct {
	db                *gorm.DB
	currenciesService currenciesService.CurrenciesServiceInterface
	logger            log15.Logger
}

func NewCardTypeRepository(
	db *gorm.DB,
	currenciesService currenciesService.CurrenciesServiceInterface,
	logger log15.Logger,
) CardTypeRepositoryInterface {
	return &cardTypeRepository{
		db:                db,
		currenciesService: currenciesService,
		logger:            logger.New("Repository", "CardTypeRepository"),
	}
}

// Returns array of records with pagination
func (repo *cardTypeRepository) FindByParams(params url.Values) ([]*model.CardType, error) {
	var cardTypes []*model.CardType
	query := paginate(repo.db, params)
	if err := query.Find(&cardTypes).Error; err != nil {
		return nil, err
	} else {
		return cardTypes, nil
	}
}

// Returns count of records
func (repo *cardTypeRepository) CountByParams(params url.Values) (count uint64, err error) {
	var cardTypes []*model.CardType
	err = repo.db.Find(&cardTypes).Count(&count).Error
	if err != nil && err.Error() == "sql: no rows in result set" {
		return 0, nil
	}
	return
}

// FindByNameAndCurrencyCode finds account type by currency code and name
func (repo *cardTypeRepository) FindByNameAndCurrencyCode(name, currencyCode string) (*model.CardType, error) {
	var cardType model.CardType

	if err := repo.db.Where("name = ? AND currency_code = ?", name, currencyCode).First(&cardType).Error; err != nil {
		return nil, err
	}

	return &cardType, nil
}

// Create creates new card type
func (repo *cardTypeRepository) Create(cardType *model.CardType) (*model.CardType, error) {
	if err := repo.db.Create(cardType).Error; err != nil {
		return nil, err
	}

	return cardType, nil
}

// Receive id and fields with name as in a struct. Updates listed fields
func (repo *cardTypeRepository) UpdateFields(id uint32, fields map[string]interface{}) (*model.CardType, error) {
	cardType := model.CardType{Id: &id}
	transformedFields := transformFieldsToDb(fields)
	if err := repo.db.Model(&cardType).Updates(transformedFields).Error; err != nil {
		return nil, err
	}
	return &cardType, nil
}

func (repo *cardTypeRepository) Get(id uint32, includes *list_params.Includes) (*model.CardType, error) {
	cardType := new(model.CardType)
	query := repo.db
	if includes != nil {
		for _, preloadName := range includes.GetPreloads() {
			query = query.Preload(preloadName)
		}
	}

	if err := query.Where("id = ?", id).First(cardType).Error; err != nil {
		return nil, err
	}
	return cardType, nil
}

func (self *cardTypeRepository) GetList(params *list_params.ListParams) (
	[]*model.CardType, error) {
	var cardTypes []*model.CardType

	str, arguments := params.GetWhereCondition()
	query := self.db.Where(str, arguments...)

	query = query.Order(params.GetOrderByString())

	query = query.Limit(params.GetLimit())
	query = query.Offset(params.GetOffset())

	query = query.Joins(params.GetJoinCondition())

	for _, preloadName := range params.GetPreloads() {
		query = query.Preload(preloadName)
	}

	if err := query.Find(&cardTypes).Error; err != nil {
		return cardTypes, err
	}

	interfaceCardTypes := make([]interface{}, len(cardTypes))
	for i, cardTypePtr := range cardTypes {
		interfaceCardTypes[i] = cardTypePtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceCardTypes); err != nil {
			return cardTypes, err
		}
	}

	return cardTypes, nil
}

func (self *cardTypeRepository) GetCount(params *list_params.ListParams) (
	int64, error) {
	var count int64
	str, arguments := params.GetWhereCondition()
	query := self.db.Where(str, arguments...)

	if err := query.Model(&model.CardType{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

func (self *cardTypeRepository) Delete(id uint32) {
	model := model.CardType{Id: &id}
	if err := self.db.Delete(&model).Error; err != nil {
		self.logger.Error("Failed to delete card type", "error", err)
		panic("Failed to delete card type")
	}
}

func (repo *cardTypeRepository) findCurrencyById(
	array []*currencyModel.Currency, id uint32) *currencyModel.Currency {
	for _, v := range array {
		if v.Id == id {
			return v
		}
	}
	return nil
}

func (repo *cardTypeRepository) isExist(array []uint32, elem uint32) bool {
	for _, v := range array {
		if v == elem {
			return true
		}
	}
	return false
}

// TODO: move to shared, add support gorm column name tags
// transformFieldsToDb receive field names as in struct.
// Map them to field name in database
func transformFieldsToDb(fields map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range fields {
		newMap[strcase.ToSnake(k)] = v
	}
	return newMap
}

// TODO: move to shared
// Forms query woth pagination. Sets default pagination if it's not passed
func paginate(query *gorm.DB, params url.Values) *gorm.DB {
	limit, err := strconv.ParseUint(params.Get("limit"), 10, 32)
	if err != nil {
		limit = DefaultLimit
	}

	query = query.Limit(uint(limit))
	offset, err := strconv.ParseUint(params.Get("offset"), 10, 32)
	if err != nil {
		offset = DefaultOffset
	}
	query = query.Offset(uint(offset))

	return query
}
