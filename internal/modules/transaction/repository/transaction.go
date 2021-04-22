package repository

import (
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-list_params/adapters"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db}
}

func (t *TransactionRepository) Create(transaction *model.Transaction) error {
	return t.db.Create(transaction).Error
}

func (t *TransactionRepository) Delete(transaction *model.Transaction) error {
	return t.db.Delete(transaction).Error
}

func (t *TransactionRepository) Updates(transaction *model.Transaction) error {
	return t.db.Model(transaction).Updates(transaction).Error
}

func (t *TransactionRepository) GetByRequestId(id uint64) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	if err := t.db.
		Where("request_id = ?", id).
		Find(&transactions).
		Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *TransactionRepository) GetById(id uint64, preload ...string) (*model.Transaction, error) {
	transaction := &model.Transaction{}
	query := t.db.Where("id = ?", id)

	for _, v := range preload {
		query = query.Preload(v)
	}
	if err := query.
		Find(&transaction).
		Error; err != nil {
		return nil, err
	}
	return transaction, nil
}

func (t *TransactionRepository) GetVisibleByRequestId(id uint64) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	if err := t.db.
		Where("request_id = ? AND is_visible = 1", id).
		Find(&transactions).
		Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *TransactionRepository) GetByRequestIdWithAccounts(id uint64) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	if err := t.db.
		Where("request_id = ?", id).
		Preload("Account").
		Preload("RevenueAccount").
		Find(&transactions).
		Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *TransactionRepository) GetByRequestIdWithAccountAndType(id uint64) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	if err := t.db.
		Where("request_id = ?", id).
		Preload("Account").
		Preload("Account.Type").
		Find(&transactions).
		Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *TransactionRepository) FindOneByRequestIdAndPurpose(id uint64, purpose string) (*model.Transaction, error) {
	transaction := &model.Transaction{}
	if err := t.db.
		Where("request_id = ? and purpose = ?", id, purpose).
		FirstOrInit(&transaction).
		Error; err != nil {
		return nil, err
	}
	return transaction, nil
}

func (t *TransactionRepository) WrapContext(db *gorm.DB) *TransactionRepository {
	return NewTransactionRepository(db)
}

func (t *TransactionRepository) FillModel(transaction *model.Transaction) error {
	return t.db.Where("id = ?", *transaction.Id).Preload("TargetDetails").First(transaction).Error
}

func (t *TransactionRepository) GetList(params *list_params.ListParams) (
	[]*model.Transaction, error,
) {
	var transactions []*model.Transaction
	adapter := adapters.NewGorm(t.db)
	err := adapter.LoadList(&transactions, params, "transactions")

	return transactions, err
}

func (t *TransactionRepository) GetListCount(params *list_params.ListParams) (
	uint64, error,
) {
	var count uint64
	str, arguments := params.GetWhereCondition()
	query := t.db.Where(str, arguments...)

	query = query.Joins(params.GetJoinCondition())

	if err := query.Model(&model.Transaction{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}
