package repository

import (
	"errors"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-list_params/adapters"
	"github.com/jinzhu/gorm"
)

var ErrorDuplicate = errors.New("fee is already assigned")

type TransferFee struct {
	db *gorm.DB
}

func NewTransferFee(db *gorm.DB) *TransferFee {
	return &TransferFee{db: db}
}

func (t *TransferFee) Create(feeModel *model.TransferFee) error {
	duplicate := &model.TransferFee{
		Name:           feeModel.Name,
		RequestSubject: feeModel.RequestSubject,
	}

	if t.db.Find(duplicate, duplicate).Error == nil {
		t.db.Rollback()
		return errcodes.CreatePublicError(errcodes.CodeDuplicateTransferFee)
	}

	return t.db.Create(feeModel).Error
}

func (t *TransferFee) AttachToUserGroup(transferFeeId, userGroupId uint64) error {
	relation := &model.TransferFeeUserGroup{
		TransferFeeId: &transferFeeId,
		UserGroupId:   &userGroupId,
	}

	err := t.db.Find(relation, relation).Error
	if err == gorm.ErrRecordNotFound {
		return t.db.
			Save(relation).
			Error
	}

	if err == nil {
		return ErrorDuplicate
	}

	return err
}

func (t *TransferFee) DetachFromUserGroup(transferFeeId, userGroupId uint64) error {
	relation := &model.TransferFeeUserGroup{
		TransferFeeId: &transferFeeId,
		UserGroupId:   &userGroupId,
	}
	return t.db.
		Delete(relation, relation).
		Error
}

func (t *TransferFee) GetAllByRequestSubject(requestSubject string) (fees []*model.TransferFee, err error) {
	err = t.db.
		Model(&model.TransferFee{}).
		Preload("Relations").
		Find(&fees, "request_subject = ?", requestSubject).
		Error
	return
}

func (t *TransferFee) Updates(feeModel *model.TransferFee) error {
	return t.db.Model(feeModel).Updates(feeModel).Error
}

func (t *TransferFee) FindById(id uint64) (*model.TransferFee, error) {
	resultModel := &model.TransferFee{}
	return resultModel, t.db.Preload("Relations").Find(resultModel, "id = ?", id).Error
}

func (t *TransferFee) Delete(feeModel *model.TransferFee) error {
	return t.db.Delete(feeModel).Error
}

func (t *TransferFee) FirstByUserGroupIdAndRequestSubject(userGroup uint64, requestSubject constants.Subject) (*model.TransferFee, error) {
	resultModel := &model.TransferFee{}
	err := t.db.
		Model(resultModel).
		Preload("Relations").
		Joins("INNER JOIN transfer_fees_user_groups tf ON tf.user_group_id = ? AND tf.transfer_fee_id = transfer_fees.id", userGroup).
		First(resultModel, "request_subject = ?", requestSubject).
		Error

	return resultModel, err
}

func (t *TransferFee) FindAllByUserGroupIdAndRequestSubject(userGroup uint64, requestSubject constants.Subject) (fees []*model.TransferFee, err error) {
	err = t.db.
		Model(&model.TransferFee{}).
		Preload("Relations").
		Joins("INNER JOIN transfer_fees_user_groups tf ON tf.user_group_id = ? AND tf.transfer_fee_id = transfer_fees.id", userGroup).
		Find(&fees, "request_subject = ?", requestSubject).
		Error

	return
}

func (t *TransferFee) GetList(params *list_params.ListParams) (fees []*model.TransferFee, err error) {
	adapter := adapters.NewGorm(t.db)
	return fees, adapter.LoadList(&fees, params, (&model.TransferFee{}).TableName())
}

func (copy TransferFee) WrapContext(db *gorm.DB) *TransferFee {
	copy.db = db
	return &copy
}
