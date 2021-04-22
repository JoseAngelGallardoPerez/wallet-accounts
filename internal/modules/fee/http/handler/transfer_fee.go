package handler

import (
	"fmt"
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	httpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/fee"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/filters"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/form"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/model"
	"github.com/Confialink/wallet-accounts/internal/modules/fee/repository"
)

type TransferFee struct {
	serviceTransferFee              *fee.ServiceTransferFee
	transferFeeRepository           *repository.TransferFee
	transferFeeParametersRepository *repository.TransferFeeParameters
	contextService                  httpService.ContextInterface
	logger                          log15.Logger
}

func NewTransferFee(
	serviceTransferFee *fee.ServiceTransferFee,
	transferFeeRepository *repository.TransferFee,
	transferFeeParametersRepository *repository.TransferFeeParameters,
	contextService httpService.ContextInterface,
	logger log15.Logger,
) *TransferFee {
	return &TransferFee{
		serviceTransferFee:              serviceTransferFee,
		transferFeeRepository:           transferFeeRepository,
		transferFeeParametersRepository: transferFeeParametersRepository,
		contextService:                  contextService,
		logger:                          logger.New("Handler", "TransferFee"),
	}
}

func (t *TransferFee) ListFees(c *gin.Context) {
	params := t.getListParams(c)
	params.AllowFilters([]string{"currencyCode", "requestSubject", "userGroupId"})

	if ok, errorsList := params.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	requestSubject := c.Param("requestSubject")
	params.AddFilter("requestSubject", []string{requestSubject})
	params.Includes.AddIncludes("Relations")

	fees, err := t.transferFeeRepository.GetList(params)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to get fees by request subject"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("requestSubject", requestSubject)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusOK, response.NewWithList(fees))
}

func (t *TransferFee) ListUserFees(c *gin.Context) {
	user := t.contextService.MustGetCurrentUser(c)
	params := t.getListParams(c)

	if ok, errorsList := params.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
	}

	requestSubject := c.Param("requestSubject")
	params.AddFilter("requestSubject", []string{requestSubject})
	params.AddFilter("userGroupId", []string{fmt.Sprintf("%d", user.GetGroupId())})

	fees, err := t.transferFeeRepository.GetList(params)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to get fees by request subject"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("requestSubject", requestSubject)
		errors.AddErrors(c, &privateError)
		return
	}
	c.JSON(http.StatusOK, response.NewWithList(fees))
}

func (t *TransferFee) GetFee(c *gin.Context) {
	id, typedError := t.contextService.GetIdParam(c)
	if typedError != nil {
		errors.AddErrors(c, typedError)
		return
	}

	fee, err := t.transferFeeRepository.FindById(id)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to get fee by id"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("id", id)
		errors.AddErrors(c, &privateError)
		return
	}
	c.JSON(http.StatusOK, response.New().SetData(fee))
}

func (t *TransferFee) DeleteFee(c *gin.Context) {
	id, typedErr := t.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	fee, err := t.transferFeeRepository.FindById(id)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to get fee by id"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("id", id)
		errors.AddErrors(c, &privateError)
		return
	}

	err = t.transferFeeRepository.Delete(fee)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to delete fee"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("id", id)
		errors.AddErrors(c, &privateError)
		return
	}

	c.Status(http.StatusOK)
}

func (t *TransferFee) ListFeeParameters(c *gin.Context) {
	feeId, typedErr := t.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	params, err := t.transferFeeParametersRepository.GetAllByTransferFeeId(feeId)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to get transfer fee parameters"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("feeId", feeId)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusOK, response.NewWithList(params))
}

func (t *TransferFee) Create(c *gin.Context) {
	transferForm := &form.TransferFee{}
	if err := c.ShouldBind(transferForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	feeModel, err := t.serviceTransferFee.Create(transferForm, nil)

	if err != nil {
		privateError := errors.PrivateError{Message: "failed to create transfer fee"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("name", *transferForm.Name)
		privateError.AddLogPair("requestSubject", *transferForm.RequestSubject)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusCreated, response.New().SetData(feeModel))
}

func (t *TransferFee) Update(c *gin.Context) {
	feeId, typedErr := t.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	updateForm := &form.UpdateTransferFee{}
	if err := c.ShouldBind(updateForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	feeModel, err := t.serviceTransferFee.Update(feeId, updateForm, nil)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to update transfer fee"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
	c.JSON(http.StatusOK, response.New().SetData(feeModel))
}

func (t *TransferFee) getListParams(c *gin.Context) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(c.Request.URL.RawQuery, model.TransferFee{})

	//params.AllowPagination()
	params.AllowFilters([]string{"currencyCode", "requestSubject"})

	params.AddCustomFilter("currencyCode", filters.CurrencyCodeEq)
	params.AddCustomFilter("userGroupId", filters.UserGroupEq)

	return params
}
