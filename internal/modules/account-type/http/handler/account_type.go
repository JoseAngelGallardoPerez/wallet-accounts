package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-model_serializer"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	serviceAccountType "github.com/Confialink/wallet-accounts/internal/modules/account-type/service"
	"github.com/Confialink/wallet-accounts/internal/modules/account/service"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"
)

// AccountHandler
type AccountTypeHandler struct {
	repo              *repository.AccountTypeRepository
	contextService    appHttpService.ContextInterface
	accountService    *service.AccountService
	systemLogsService *system_logs.SystemLogsService
	logger            log15.Logger
	service           *serviceAccountType.AccountTypeService
}

// AccountTypeHandlerInterface is interface for handler functionality
// that ought to be implemented manually.
type AccountTypeHandlerInterface interface {
	ListHandler(c *gin.Context)
	GetHandler(c *gin.Context)
	CreateHandler(c *gin.Context)
	UpdateHandler(c *gin.Context)
	DeleteHandler(c *gin.Context)
	NotFoundHandler(c *gin.Context)
}

var periodFields = []interface{}{"ID", "Name"}
var currencyFields = []interface{}{"Id", "Code"}

var showListOutputFields = []interface{}{
	"ID",
	"Name",
	"CurrencyCode",
	"DepositAnnualInterestRate",
	"CreditAnnualInterestRate",
	"DepositPayoutPeriodID",
	"AutoNumberGeneration",
	"NumberPrefix",
	map[string][]interface{}{"DepositPayoutPeriod": periodFields},
	map[string][]interface{}{"CreditChargePeriod": periodFields},
	map[string][]interface{}{"Currency": currencyFields},
}

// NewAccountService creates new account service
func NewAccountTypeHandler(repo *repository.AccountTypeRepository,
	contextService appHttpService.ContextInterface,
	accountService *service.AccountService,
	service *serviceAccountType.AccountTypeService,
	systemLogsService *system_logs.SystemLogsService,
	logger log15.Logger,
) *AccountTypeHandler {
	return &AccountTypeHandler{
		repo:              repo,
		contextService:    contextService,
		accountService:    accountService,
		systemLogsService: systemLogsService,
		service:           service,
		logger:            logger.New("Handler", "account_type.AccountTypeHandler"),
	}
}

// ListHandler returns the list of account types
// @TODO: move logic to a service
func (h *AccountTypeHandler) ListHandler(c *gin.Context) {
	listParams := h.getListParams(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	items, err := h.repo.GetList(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of account types"}
		privateError.AddLogPair("error", err)
		return
	}
	count, countErr := h.repo.GetListCount(listParams)
	if countErr != nil {
		privateError := errors.PrivateError{Message: "can't retrieve count of account types"}
		privateError.AddLogPair("error", err)
		return
	}

	serialized := make([]interface{}, len(items))
	fields := listParams.GetOutputFields()
	for i, v := range items {
		serialized[i] = model_serializer.Serialize(v, fields)
	}
	r := response.NewWithListAndPageLinks(serialized, count, c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, r)
}

// GetHandler returns account type by id
func (h *AccountTypeHandler) GetHandler(c *gin.Context) {
	logger := h.logger.New("action", "GetHandler")

	id := h.getIdParam(c)

	accountType, err := h.repo.FindByID(id)

	if err != nil {
		logger.Error("can't retrieve account type", "err", err, "account type id", id)
		errcodes.AddError(c, errcodes.CodeAccountTypeNotFound)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(accountType))
}

// CreateHandler creates a new account type
// @TODO: move logic to a service
func (h *AccountTypeHandler) CreateHandler(c *gin.Context) {
	currentUser := h.contextService.MustGetCurrentUser(c)
	var public model.AccountTypePublic

	if err := c.ShouldBindJSON(&public); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	publicJson, err := json.Marshal(public)

	if nil != err {
		privateError := errors.PrivateError{Message: "can't marshal json"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	var accountType *model.AccountType
	_ = json.Unmarshal(publicJson, &accountType)

	createdAccountType, err := h.service.Create(accountType)

	if nil != err {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	h.systemLogsService.LogCreateAccountTypeAsync(createdAccountType, currentUser.UID)

	c.JSON(http.StatusCreated, response.New().SetData(createdAccountType))
}

// UpdateHandler updates an account type
// @TODO: move logic to a service
func (h *AccountTypeHandler) UpdateHandler(c *gin.Context) {
	logger := h.logger.New("action", "UpdateHandler")

	currentUser := h.contextService.MustGetCurrentUser(c)

	var editable model.AccountTypeEditable

	if err := c.ShouldBindJSON(&editable); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	editableJson, err := json.Marshal(editable)

	if nil != err {
		privateError := errors.PrivateError{Message: "can't marshal json"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	id := h.getIdParam(c)
	accountType, err := h.repo.FindByID(id)

	if nil != err {
		logger.Error("can't retrieve account type", "err", err, "account type id", id)
		errcodes.AddError(c, errcodes.CodeAccountTypeNotFound)
		return
	}

	old, _ := h.repo.FindByID(id) // @TODO: implement clone instead
	json.Unmarshal(editableJson, &accountType)
	updatedAccountType, err := h.service.Update(accountType)

	if nil != err {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	h.systemLogsService.LogModifyAccountTypeAsync(old, accountType, currentUser.UID)

	c.JSON(http.StatusOK, response.New().SetData(updatedAccountType))
}

// DeleteHandler delete an account type
// @TODO: move logic to a service
func (h *AccountTypeHandler) DeleteHandler(c *gin.Context) {
	logger := h.logger.New("action", "DeleteHandler")
	id := h.getIdParam(c)
	accountType, err := h.repo.FindByID(id)

	if nil != err {
		logger.Error("can't retrieve account type", "err", err, "account type id", id)
		errcodes.AddError(c, errcodes.CodeAccountTypeNotFound)
		return
	}

	count, _ := h.accountService.CountByAccountTypeId(accountType.ID)

	if count > 0 {
		logger.Error("can't delete account type, account type contains accounts", "account type id", id)
		errcodes.AddError(c, errcodes.CodeAccountTypeContainsAccounts)
		return
	}

	err = h.repo.Delete(accountType)

	if err != nil {
		privateError := errors.PrivateError{Message: "can't delete account type"}
		privateError.AddLogPair("error", err)
		privateError.AddLogPair("account type id", id)
		errors.AddErrors(c, &privateError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AccountTypeHandler) NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
}

// getIdParam returns id or nil
func (h *AccountTypeHandler) getIdParam(c *gin.Context) uint64 {
	id := c.Params.ByName("id")

	// convert string to uint
	id64, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewWithError(
			"id param must be an integer",
			err.Error(),
			http.StatusBadRequest,
			nil,
		))
		return 0
	}

	return uint64(id64)
}

func (self *AccountTypeHandler) getListParams(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.AccountType{})
	params.AllowSelectFields(showListOutputFields)
	params.AllowPagination()
	self.addIncludes(params)
	self.addSortings(params)
	self.addFilters(params)
	return params
}

func (self *AccountTypeHandler) addFilters(params *list_params.ListParams) {
	params.AllowFilters([]string{"name"})
	params.AddCustomFilter("name", self.nameFilter)
}

func (self *AccountTypeHandler) addIncludes(params *list_params.ListParams) {
	params.AllowIncludes([]string{"depositPayoutPeriod", "creditChargePeriod"})
}

func (self *AccountTypeHandler) addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"id", "name", "currencyCode", "annualInterestRate", "period.name"})
	params.AddCustomSortings("annualInterestRate", self.annualInterestRateSorting)
	params.AddCustomSortings("period.name", self.periodNameSorting)
}

func (self *AccountTypeHandler) annualInterestRateSorting(direction string,
	_ *list_params.ListParams) (string, error,
) {
	return fmt.Sprintf(`CASE
	 WHEN credit_annual_interest_rate IS NULL THEN deposit_annual_interest_rate
	 ELSE credit_annual_interest_rate
	 END %s
	`, direction), nil
}

func (self *AccountTypeHandler) periodNameSorting(direction string,
	params *list_params.ListParams) (string, error,
) {
	params.AddLeftJoin("payment_periods AS credit_periods", "account_types.credit_charge_period_id = credit_periods.id")
	params.AddLeftJoin("payment_periods AS deposit_periods", "account_types.deposit_payout_period_id = deposit_periods.id")
	return fmt.Sprintf(`CASE
	WHEN credit_annual_interest_rate IS NULL THEN deposit_periods.name
	ELSE credit_periods.name
	END %s
	`, direction), nil
}

func (self *AccountTypeHandler) nameFilter(inputValues []string,
	_ *list_params.ListParams) (dbConditionPart string, dbValues interface{}) {
	return "name LIKE ?", fmt.Sprintf("%%%s%%", inputValues[0])
}
