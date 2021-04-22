package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card/model"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/Confialink/wallet-pkg-list_params"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/iancoleman/strcase"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

type CardRepositoryInterface interface {
	Create(*model.Card) (*model.Card, error)
	Get(uint32, *list_params.Includes) (*model.Card, error)
	GetByNumber(number string, includes *list_params.Includes) (card *model.Card, err error)
	UpdateFields(uint32, map[string]interface{}) (*model.Card, error)
	GetList(*list_params.ListParams) ([]*model.Card, error)
	GetListByCardTypeId(id uint32) []*model.Card
	GetListCount(*list_params.ListParams) (uint64, error)
	FillUsers(cards []*model.Card) error
	BulkCreate(cards []*model.Card) ([]*model.Card, error)
	WrapContext(db *gorm.DB) CardRepositoryInterface
}

type cardRepository struct {
	db          *gorm.DB
	userService *service.UserService
	logger      log15.Logger
}

func NewCardRepository(
	db *gorm.DB,
	userService *service.UserService,
	logger log15.Logger,
) CardRepositoryInterface {
	return &cardRepository{db: db, userService: userService, logger: logger.New("Repository", "CardRepository")}
}

// Create takes all fields in passed model to fill database
func (c *cardRepository) Create(card *model.Card) (*model.Card, error) {
	if err := c.db.Create(card).Error; err != nil {
		return nil, err
	}
	return card, nil
}

// Get returns model by id
func (c *cardRepository) Get(id uint32, includes *list_params.Includes) (card *model.Card, err error) {
	card = new(model.Card)
	query := c.db
	if includes != nil {
		for _, preloadName := range includes.GetPreloads() {
			query = query.Preload(preloadName)
		}
	}

	err = query.Where("id = ?", id).First(card).Error
	if err != nil {
		return
	}

	if includes != nil {
		interfaceCards := []interface{}{card}
		for _, customIncludesFunc := range includes.GetCustomIncludesFunctions() {
			if err := customIncludesFunc(interfaceCards); err != nil {
				return nil, err
			}
		}
	}

	return
}

// Get returns model by id
func (c *cardRepository) GetByNumber(number string, includes *list_params.Includes) (*model.Card, error) {
	var card model.Card
	query := c.db
	if includes != nil {
		for _, preloadName := range includes.GetPreloads() {
			query = query.Preload(preloadName)
		}
	}

	err := query.Where("number = ?", number).First(&card).Error
	if err != nil {
		return nil, err
	}

	if includes != nil {
		interfaceCards := []interface{}{card}
		for _, customIncludesFunc := range includes.GetCustomIncludesFunctions() {
			if err := customIncludesFunc(interfaceCards); err != nil {
				return nil, err
			}
		}
	}

	return &card, nil
}

// Receive id and fields with name as in a struct. Updates listed fields
func (c *cardRepository) UpdateFields(
	id uint32, fields map[string]interface{}) (*model.Card, error) {
	card := model.Card{Id: &id}
	transformedFields := transformFieldsToDb(fields)
	if err := c.db.Model(&card).Updates(transformedFields).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

func (c *cardRepository) GetList(params *list_params.ListParams) ([]*model.Card, error) {
	var cards []*model.Card

	str, arguments := params.GetWhereCondition()
	query := c.db.Where(str, arguments...)

	query = query.Order(params.GetOrderByString())

	if params.GetLimit() != 0 {
		query = query.Limit(params.GetLimit())
	}
	query = query.Offset(params.GetOffset())

	query = query.Joins(params.GetJoinCondition())

	for _, preloadName := range params.GetPreloads() {
		query = query.Preload(preloadName)
	}

	if err := query.Find(&cards).Error; err != nil {
		return cards, err
	}

	interfaceCards := make([]interface{}, len(cards))
	for i, cardPtr := range cards {
		interfaceCards[i] = cardPtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceCards); err != nil {
			return cards, err
		}
	}

	return cards, nil
}

func (c *cardRepository) GetListCount(params *list_params.ListParams) (uint64, error) {
	var count uint64
	str, arguments := params.GetWhereCondition()
	query := c.db.Where(str, arguments...)

	query = query.Joins(params.GetJoinCondition())

	if err := query.Model(&model.Card{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

// BulkCreate creates list of cards
func (c *cardRepository) BulkCreate(cards []*model.Card) ([]*model.Card, error) {
	tx := c.db.Begin()
	for _, card := range cards {
		if err := tx.Create(card).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return cards, nil
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

func (c *cardRepository) FillUsers(cards []*model.Card) error {
	uids := make([]string, 0)
	for _, v := range cards {
		if !c.isExist(uids, *v.UserId) {
			uids = append(uids, *v.UserId)
		}
	}

	users, err := c.userService.GetByUIDs(uids)
	if err != nil {
		return err
	}

	for _, v := range cards {
		user := c.findUserByUID(users, *v.UserId)
		if user != nil {
			c.fillUser(v, user)
		}
	}

	return nil
}

func (c *cardRepository) GetListByCardTypeId(id uint32) []*model.Card {
	var cards []*model.Card
	err := c.db.Where("card_type_id = ?", id).Find(&cards).Error
	if err != nil {
		c.logger.Error("Failed to get cards by card type id", "err", err)
		panic("Failed to get cards by card type id")
	}
	return cards
}

func (c cardRepository) WrapContext(db *gorm.DB) CardRepositoryInterface {
	c.db = db
	return &c
}

func (c *cardRepository) isExist(array []string, elem string) bool {
	for _, v := range array {
		if v == elem {
			return true
		}
	}
	return false
}

func (c *cardRepository) findUserByUID(
	array []*userpb.User, uid string,
) *userpb.User {
	for _, v := range array {
		if v.UID == uid {
			return v
		}
	}
	return nil
}

func (c *cardRepository) fillUser(
	card *model.Card, user *userpb.User,
) {
	card.User = &model.User{
		Id:        &user.UID,
		Username:  &user.Username,
		Email:     &user.Email,
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
	}
}
