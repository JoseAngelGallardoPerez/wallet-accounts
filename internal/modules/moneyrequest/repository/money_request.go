package repository

import (
	list_params "github.com/Confialink/wallet-pkg-list_params"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/model"
)

type MoneyRequest struct {
	db *gorm.DB
}

// NewMoneyRequest return new MoneyRequest repository
func NewMoneyRequest(db *gorm.DB) *MoneyRequest {
	return &MoneyRequest{
		db: db,
	}
}

// Create create new Invoice
func (r *MoneyRequest) Create(moneyRequest *model.MoneyRequest) error {
	return r.db.Create(moneyRequest).Error
}

func (r *MoneyRequest) Update(moneyRequest *model.MoneyRequest) error {
	return r.db.Save(moneyRequest).Error
}

func (r *MoneyRequest) GetByTargetUID(id uint64, targetUID string) (*model.MoneyRequest, error) {
	var result model.MoneyRequest

	err := r.db.
		Where("id = ?", id).
		Where("target_user_id = ?", targetUID).
		First(&result).Error
	return &result, err
}

// GetList returns records from passed ListParams
func (r *MoneyRequest) GetList(params *list_params.ListParams) (
	[]*model.MoneyRequest, error) {
	var records []*model.MoneyRequest

	str, arguments := params.GetWhereCondition()
	query := r.db.Where(str, arguments...)

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

	if err := query.Find(&records).Error; err != nil {
		return records, err
	}

	interfaceRecords := make([]interface{}, len(records))
	for i, accountPtr := range records {
		interfaceRecords[i] = accountPtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceRecords); err != nil {
			return records, err
		}
	}

	return records, nil
}

func (r *MoneyRequest) GetCount(params *list_params.ListParams) (int64, error) {
	var count int64
	str, arguments := params.GetWhereCondition()
	query := r.db.Where(str, arguments...)

	query = query.Joins(params.GetJoinCondition())

	if err := query.Model(&model.MoneyRequest{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

func (*MoneyRequest) WrapContext(db *gorm.DB) *MoneyRequest {
	return NewMoneyRequest(db)
}
