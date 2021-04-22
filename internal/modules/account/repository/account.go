package repository

import (
	"errors"
	"net/url"
	"strconv"

	list_params "github.com/Confialink/wallet-pkg-list_params"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	userModel "github.com/Confialink/wallet-accounts/internal/modules/user/model"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
)

// AccountRepository is user repository for CRUD operations.
type AccountRepository struct {
	db          *gorm.DB
	userService *service.UserService
}

// NewAccountRepository creates new repository
func NewAccountRepository(db *gorm.DB, userService *service.UserService) *AccountRepository {
	return &AccountRepository{db, userService}
}

// GetList returns records from passed ListParams
func (a *AccountRepository) GetList(params *list_params.ListParams) (
	[]*model.Account, error) {
	var accounts []*model.Account

	str, arguments := params.GetWhereCondition()
	query := a.db.Where(str, arguments...)

	query = query.Order(params.GetOrderByString())

	if params.GetLimit() != 0 {
		query = query.Limit(params.GetLimit())
	}
	query = query.Offset(params.GetOffset())

	query = query.Joins(params.GetJoinCondition())

	for _, preloadName := range params.GetPreloads() {
		query = query.Preload(preloadName)
	}

	groupBy := params.GetGroupBy()
	if groupBy != nil {
		query = query.Group(*groupBy)
	}

	if err := query.Find(&accounts).Error; err != nil {
		return accounts, err
	}

	interfaceAccounts := make([]interface{}, len(accounts))
	for i, accountPtr := range accounts {
		interfaceAccounts[i] = accountPtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceAccounts); err != nil {
			return accounts, err
		}
	}

	return accounts, nil
}

func (a *AccountRepository) GetCount(params *list_params.ListParams) (int64, error) {
	var count int64
	str, arguments := params.GetWhereCondition()
	query := a.db.Where(str, arguments...)

	query = query.Joins(params.GetJoinCondition())

	if err := query.Model(&model.Account{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

// FindByID find account by id
func (a *AccountRepository) FindByID(id uint64, preloads ...string) (*model.Account, error) {
	var account model.Account
	account.ID = id
	db := a.db
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	db = db.Preload("Type").
		Preload("InterestAccount")
	if err := db.First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// FindByIDAndUserID find account by id and user id
func (a *AccountRepository) FindByIDAndUserID(id uint64, userID string, preloads ...string) (*model.Account, error) {
	var account model.Account
	account.ID = id
	account.UserId = userID
	db := a.db
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	db = db.Preload("Type").
		Preload("InterestAccount")
	if err := db.First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (a *AccountRepository) FindManyByIds(ids []uint64, preloads ...string) ([]*model.Account, error) {
	var accounts []*model.Account
	db := a.db
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Where("id IN (?)", ids).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// FindByNumber find account by number
func (a *AccountRepository) FindByNumber(number string) (*model.Account, error) {
	var account model.Account
	if err := a.db.Preload("Type").Where("number = ?", number).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// CountByAccountTypeId count accounts by account type id
func (a *AccountRepository) CountByAccountTypeId(id uint64) (uint64, error) {
	var account []*model.Account
	var count uint64

	if err := a.db.
		Where("type_id = ?", id).
		Model(&account).
		Count(&count).
		Error; err != nil {
		return count, err
	}

	return count, nil
}

// CountByTypeIdAndUserId count accounts by account type id and user id
func (a *AccountRepository) CountByTypeIdAndUserId(typeId uint64, userId string) (uint64, error) {
	var account []*model.Account
	var count uint64

	if err := a.db.
		Where("type_id = ? AND user_id = ?", typeId, userId).
		Model(&account).
		Count(&count).
		Error; err != nil {
		return count, err
	}

	return count, nil
}

// Create creates new account
func (a *AccountRepository) Create(account *model.Account) (*model.Account, error) {
	if err := a.db.Create(account).Error; err != nil {
		return nil, err
	}
	return a.FindByID(account.ID)
}

// BulkCreate creates many accounts
func (a *AccountRepository) BulkCreate(accounts []*model.Account) ([]*model.Account, error) {
	tx := a.db.Begin()
	for _, account := range accounts {
		if err := tx.Create(account).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return accounts, nil
}

// Update updates an existing account
func (a *AccountRepository) Update(account *model.Account) (*model.Account, error) {
	if err := a.db.Save(account).Error; err != nil {
		return nil, err
	}
	return a.FindByID(account.ID)
}

// Updates updates only not default values of an existing account
func (a *AccountRepository) Updates(account *model.Account) error {
	return a.db.Model(account).Updates(account).Error
}

// Delete delete an existing account
func (a *AccountRepository) Delete(account *model.Account) error {
	if err := a.db.Delete(account).Error; err != nil {
		return err
	}
	return nil
}

// UpdateCreditAvailableAmount updates available amount value using diff value
func (a *AccountRepository) UpdateAvailableAmountByAccountTypeId(diff decimal.Decimal, accountTypeId uint64) error {
	table := (&model.Account{}).TableName()
	return a.db.Exec("UPDATE "+table+" SET available_amount = available_amount + ? WHERE type_id = ?", diff, accountTypeId).Error
}

//TODO: Do not use DB for this
// Generate account number
func (a *AccountRepository) GenerateAccountNumberWithPrefix(prefix *string) string {
	type Result struct {
		GeneratedNum string
	}

	var account model.Account
	var result Result

	s := "FLOOR(RAND() * 999999999)"
	if prefix != nil {
		s = "CONCAT('" + *prefix + "', FLOOR(RAND() * 999999999))"
	}

	a.db.Raw("SELECT " + s + " AS generated_num " +
		"FROM (SELECT FLOOR(RAND() * 999999999) AS random_num) AS numbers " +
		"WHERE 'generated_num' NOT IN (SELECT number FROM " + account.TableName() + ") " +
		"LIMIT 1").Scan(&result)

	return result.GeneratedNum
}

// GetBalance return account's balance
func (a *AccountRepository) GetBalance(id uint64) (decimal.Decimal, error) {
	type Result struct {
		Balance decimal.Decimal
	}

	var result Result
	var account model.Account
	if err := a.db.Table(account.TableName()).
		Select("balance").
		Where("id = ?", id).
		Scan(&result).
		Error; nil != err {
		return decimal.Zero, err
	}
	return result.Balance, nil
}

// GetAvailableAmount return account's available amount
func (a *AccountRepository) GetAvailableAmount(id uint64) (decimal.Decimal, error) {
	type Result struct {
		AvailableAmount decimal.Decimal
	}

	var result Result
	var account model.Account
	if err := a.db.Table(account.TableName()).
		Select("available_amount").
		Where("id = ?", id).
		Scan(&result).
		Error; nil != err {
		return decimal.Zero, err
	}
	return result.AvailableAmount, nil
}

func (a *AccountRepository) WrapContext(db *gorm.DB) *AccountRepository {
	return NewAccountRepository(db, a.userService)
}

func (a *AccountRepository) FillUsers(accounts []*model.Account) error {
	uids := make([]string, 0)
	for _, v := range accounts {
		if !a.isExist(uids, v.UserId) {
			uids = append(uids, v.UserId)
		}
	}

	users, err := a.userService.GetByUIDs(uids)
	if err != nil {
		return err
	}

	for _, v := range accounts {
		user := a.findUserByUID(users, v.UserId)
		if user != nil {
			a.fillUser(v, user)
		}
	}

	return nil
}

func (a *AccountRepository) FillUser(account *model.Account) error {
	if nil == account {
		return errors.New("empty value provided")
	}

	user, err := a.userService.GetByUID(account.UserId)
	if err != nil {
		return err
	}

	if user != nil {
		a.fillUser(account, user)
	}

	return nil
}

func (a *AccountRepository) isExist(array []string, elem string) bool {
	for _, v := range array {
		if v == elem {
			return true
		}
	}
	return false
}

func (a *AccountRepository) findUserByUID(
	array []*userpb.User, uid string,
) *userpb.User {
	for _, v := range array {
		if v.UID == uid {
			return v
		}
	}
	return nil
}

func (a *AccountRepository) fillUser(
	account *model.Account, user *userpb.User,
) {
	account.User = &userModel.User{
		UID:       &user.UID,
		Email:     &user.Email,
		Username:  &user.Username,
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		RoleName:  &user.RoleName,
		GroupId:   &user.GroupId,
	}
}

// paginate check if query parameters limit, after and offset are set
// and applies them to query builder
// argument usedLimit will be set the same value as for limit
// it is used in order to determine whether there are more records exist
// limit always increments by 1
func (a *AccountRepository) paginate(query *gorm.DB, params url.Values) *gorm.DB {
	limit, err := strconv.ParseUint(params.Get("limit"), 10, 32)
	if nil == err {
		if limit > 100 {
			limit = 100
		}

	} else {
		limit = 15
	}

	query = query.Limit(uint(limit))
	offset, err := strconv.ParseUint(params.Get("offset"), 10, 32)
	if nil == err {
		query = query.Offset(uint(offset))
	}

	return query
}

// sliceContains checks string exists in slice
func (a *AccountRepository) sliceContains(slice []string, string string) bool {
	for _, el := range slice {
		if el == string {
			return true
		}
	}
	return false
}
