package fee

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/form"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/model"
	feeRepository "github.com/Confialink/wallet-accounts/internal/modules/fee/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

type ServiceTransferFee struct {
	db                              *gorm.DB
	transferFeeRepository           *feeRepository.TransferFee
	transferFeeParametersRepository *feeRepository.TransferFeeParameters
	logger                          log15.Logger
}

type ParamsQuery struct {
	RequestSubject constants.Subject
	CurrencyCode   string
	UserGroupId    uint64
	FeeId          *uint64
}

func NewServiceTransferFee(
	db *gorm.DB,
	transferFeeRepository *feeRepository.TransferFee,
	transferFeeParametersRepository *feeRepository.TransferFeeParameters,
	logger log15.Logger,
) *ServiceTransferFee {
	return &ServiceTransferFee{
		db:                              db,
		transferFeeRepository:           transferFeeRepository,
		transferFeeParametersRepository: transferFeeParametersRepository,
		logger:                          logger.New("module", "fee", "service", "transferFee"),
	}
}

func (s *ServiceTransferFee) Create(createForm *form.TransferFee, tx *gorm.DB) (*model.TransferFee, error) {
	if tx == nil {
		tx = s.db.Begin()
	}

	name := createForm.Name
	userGroupIds := createForm.UserGroupIds
	params := createForm.Parameters

	logger := s.logger.New("func", "Create")
	feeRepo := s.transferFeeRepository.WrapContext(tx)

	feeModel := &model.TransferFee{
		Name:           createForm.Name,
		RequestSubject: createForm.RequestSubject,
		Relations:      make([]*model.TransferFeeUserGroup, len(createForm.UserGroupIds)),
	}

	for i, userGroupId := range userGroupIds {
		//copy value for further usage by pointer
		var groupId = userGroupId
		feeModel.Relations[i] = &model.TransferFeeUserGroup{
			UserGroupId: &groupId,
		}
	}

	err := feeRepo.Create(feeModel)

	if err != nil {
		logger.Error(
			"failed to create new fee",
			"error", err,
			"feeName", name,
		)
		tx.Rollback()
		return nil, err
	}

	paramsRepo := s.transferFeeParametersRepository.WrapContext(tx)
	for _, param := range params {
		paramsModel, err := param.ToModel()
		if err != nil {
			logger.Error("failed to get model", "error", err, "formParams", param)
			tx.Rollback()
			return nil, err
		}

		paramsModel.TransferFeeId = feeModel.Id
		err = paramsRepo.CreateUpdate(paramsModel)
		if err != nil {
			logger.Error("failed to create fee parameters", "error", err)
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()
	return feeModel, nil
}

func (s *ServiceTransferFee) Update(feeId uint64, updateForm *form.UpdateTransferFee, tx *gorm.DB) (*model.TransferFee, error) {
	if tx == nil {
		tx = s.db.Begin()
	}

	logger := s.logger.New("func", "Update")
	feeRepo := s.transferFeeRepository.WrapContext(tx)
	params := updateForm.Parameters

	feeModel, err := feeRepo.FindById(feeId)
	if err != nil {
		logger.Error("failed to find transfer fee by id", "error", err, "id", feeId)
		tx.Rollback()
		return nil, err
	}

	if updateForm.Name != nil && *updateForm.Name != *feeModel.Name {
		feeModel.Name = updateForm.Name
		err := feeRepo.Updates(&model.TransferFee{
			Id:   feeModel.Id,
			Name: updateForm.Name,
		})
		if err != nil {
			logger.Error("failed to update fee name", "error", err, "id", feeId, "newName", *updateForm.Name)
			return nil, err
		}
	}

	paramsRepo := s.transferFeeParametersRepository.WrapContext(tx)
	for _, param := range params {
		paramsModel, err := param.ToModel()
		if err != nil {
			logger.Error("failed to get model", "error", err, "formParams", param)
			tx.Rollback()
			return nil, err
		}

		paramsModel.TransferFeeId = feeModel.Id
		if param.Delete != nil && *param.Delete {
			err = paramsRepo.Delete(paramsModel)
		} else {
			err = paramsRepo.CreateUpdate(paramsModel)
		}
		if err != nil {
			logger.Error("failed to update fee parameters", "error", err)
			tx.Rollback()
			return nil, err
		}
	}

	for _, relation := range updateForm.Relations {
		if *relation.Attached {
			err = feeRepo.AttachToUserGroup(feeId, *relation.UserGroupId)
			if err == feeRepository.ErrorDuplicate {
				err = nil
			}
		} else {
			err = feeRepo.DetachFromUserGroup(feeId, *relation.UserGroupId)
		}
		if err != nil {
			logger.Error("failed to apply relation state", "error", err)
			tx.Rollback()
			return nil, err
		}
	}

	//reload model in order to reflect actual values
	feeModel, err = feeRepo.FindById(feeId)
	if err != nil {
		logger.Error("failed to find transfer fee by id", "error", err, "id", feeId)
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return feeModel, nil
}

func (s *ServiceTransferFee) FindParams(paramsQuery *ParamsQuery, tx *gorm.DB) (*model.TransferFeeParameters, error) {
	if tx == nil {
		tx = s.db
	}
	var err error

	userGroupId := paramsQuery.UserGroupId
	requestSubject := paramsQuery.RequestSubject
	currencyCode := paramsQuery.CurrencyCode

	logger := s.logger.New("func", "FindParams")

	feeRepo := s.transferFeeRepository.WrapContext(tx)

	var transferFeeModel *model.TransferFee

	if paramsQuery.FeeId != nil {
		transferFeeModel, err = feeRepo.FindById(*paramsQuery.FeeId)
		// in case if feeId is provided we have to verify if the found data matches request subject
		if err == nil {
			if paramsQuery.RequestSubject != "" && !paramsQuery.RequestSubject.EqualsTo(*transferFeeModel.RequestSubject) {
				logger.Warn("given request subject does not match fee model", "paramsQuery.RequestSubject", paramsQuery.RequestSubject, "*transferFeeModel.RequestSubject", *transferFeeModel.RequestSubject)
				return nil, errcodes.CreatePublicError(errcodes.CodeForbidden)
			}
			//if paramsQuery.UserGroupId !=
		}
	} else {
		transferFeeModel, err = feeRepo.FirstByUserGroupIdAndRequestSubject(userGroupId, requestSubject)
	}

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.Error("failed to find transfer fee", "error", err, "requestSubject", requestSubject, "userGroupId", userGroupId)
		}
		return nil, err
	}

	relatedWithGroup := false
	for _, relation := range transferFeeModel.Relations {
		if *relation.UserGroupId == userGroupId {
			relatedWithGroup = true
			break
		}
	}
	if !relatedWithGroup {
		logger.Error("transfer fee is not related to provided user group", "feeId", transferFeeModel.Id, "userGroup", userGroupId)
		err := errcodes.CreatePublicError(errcodes.CodeForbidden)
		err.Details = "invalid transfer fee"

		return nil, err
	}

	if transferFeeModel.Id == nil {
		return &model.TransferFeeParameters{}, nil
	}

	paramsRepo := s.transferFeeParametersRepository.WrapContext(tx)

	feeParamsModel, err := paramsRepo.FindByTransferFeeIdAndCurrencyCode(*transferFeeModel.Id, currencyCode)
	if err != nil {
		logger.Error("failed to find fee parameters", "error", err, "transferFeeId", *feeParamsModel.Id, "currencyCode", currencyCode)
		return nil, err
	}

	return feeParamsModel, nil
}
