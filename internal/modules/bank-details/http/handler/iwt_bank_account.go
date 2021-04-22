package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	errors "github.com/Confialink/wallet-pkg-errors"
	list_params "github.com/Confialink/wallet-pkg-list_params"
	model_serializer "github.com/Confialink/wallet-pkg-model_serializer"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepo "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	appValidator "github.com/Confialink/wallet-accounts/internal/modules/app/validator"
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/service"
)

// IwtBankAccountHandler
type IwtBankAccountHandler struct {
	repo           *repository.IwtBankAccountRepository
	accountRepo    *accountRepo.AccountRepository
	contextService appHttpService.ContextInterface
	validator      appValidator.Interface
	service        *service.IwtBankAccountService
	pdf            *service.Pdf
	logger         log15.Logger
}

// NewIwtBankAccountService creates new account service
func NewIwtBankAccountHandler(
	repo *repository.IwtBankAccountRepository,
	accountRepo *accountRepo.AccountRepository,
	contextService appHttpService.ContextInterface,
	validator appValidator.Interface,
	service *service.IwtBankAccountService,
	pdf *service.Pdf,
	logger log15.Logger,
) *IwtBankAccountHandler {
	return &IwtBankAccountHandler{
		repo:           repo,
		accountRepo:    accountRepo,
		contextService: contextService,
		validator:      validator,
		service:        service,
		pdf:            pdf,
		logger:         logger.New("Handler", "IwtBankAccountHandler"),
	}
}

var bankDetailsFields = []interface{}{
	"ID",
	"CreatedAt",
	"SwiftCode",
	"BankName",
	"Address",
	"Location",
	"Country",
	"CountryId",
	"AbaNumber",
	"Iban",
}
var beneficiaryCustomerFields = []interface{}{
	"ID",
	"CreatedAt",
	"AccountName",
	"Address",
	"Iban",
}

var showListOutputFields = []interface{}{
	"ID",
	"CurrencyCode",
	"IsIwtEnabled",
	"AdditionalInstructions",
	"BeneficiaryBankDetailsId",
	"BeneficiaryCustomerId",
	"IntermediaryBankDetailsId",
	"AdditionalInstructions",
	map[string][]interface{}{"BeneficiaryBankDetails": bankDetailsFields},
	map[string][]interface{}{"IntermediaryBankDetails": bankDetailsFields},
	map[string][]interface{}{"BeneficiaryCustomer": beneficiaryCustomerFields},
}

// ListHandler returns the list of accounts for admins
func (h *IwtBankAccountHandler) ListHandler(c *gin.Context) {
	listParams := h.getListParams(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	items, err := h.repo.GetList(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of of IWT bank details"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
	count, countErr := h.repo.GetListCount(listParams)
	if countErr != nil {
		privateError := errors.PrivateError{Message: "can't retrieve count of IWT bank details"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
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

// GetHandler returns account by id
func (h *IwtBankAccountHandler) GetHandler(c *gin.Context) {
	logger := h.logger.New("action", "GetHandler")

	id := h.getIdParam(c)
	account, err := h.repo.FindByID(id)

	if err != nil {
		logger.Error("can't retrieve iwt bank account", "err", err, "iwt bank account id", id)
		errcodes.AddError(c, errcodes.CodeIwtBankDetailsNotFound)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(account))
}

// GetByAccountIdHandler returns iwt account by account; id
func (h *IwtBankAccountHandler) GetByAccountIdHandler(c *gin.Context) {
	logger := h.logger.New("action", "GetByAccountIdHandler")

	id := h.getIdParam(c)
	account, err := h.accountRepo.FindByID(id)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
	}

	currencyCode := account.Type.CurrencyCode
	iwtAccounts, err := h.repo.FindEnabledByCurrencyCode(currencyCode)

	if err != nil {
		logger.Error("can't retrieve iwt bank account", "err", err, "iwt bank account id", id)
		errcodes.AddError(c, errcodes.CodeIwtBankDetailsNotFound)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(iwtAccounts))
}

// CreateHandler creates a new account
func (h *IwtBankAccountHandler) CreateHandler(c *gin.Context) {
	currentUser := h.contextService.MustGetCurrentUser(c)
	var public model.IwtBankAccountPublic

	err := c.ShouldBindJSON(&public)

	if err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	publicJson, err := json.Marshal(public)

	if nil != err {
		privateError := errors.PrivateError{Message: "can't marshal json"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	account := model.IwtBankAccountModel{}
	json.Unmarshal(publicJson, &account)

	createdIwtBankAccount, err := h.service.Create(&account, currentUser)

	if nil != err {
		privateError := errors.PrivateError{Message: "can't create iwt bank account"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("request data", publicJson)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusCreated, response.New().SetData(createdIwtBankAccount))
}

// UpdateHandler updates an account
func (h *IwtBankAccountHandler) UpdateHandler(c *gin.Context) {
	logger := h.logger.New("action", "UpdateHandler")
	currentUser := h.contextService.MustGetCurrentUser(c)

	editable := model.IwtBankAccountPublic{}

	err := c.ShouldBindJSON(&editable)

	data := &model.IwtBankAccountModel{
		IwtBankAccountPublic: editable,
	}

	if err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	id := h.getIdParam(c)
	account, err := h.repo.FindByID(id)

	if nil != err {
		logger.Error("can't retrieve iwt bank account", "err", err, "iwt bank account id", id)
		errcodes.AddError(c, errcodes.CodeIwtBankDetailsNotFound)
		return
	}

	updatedIwtBankAccount, err := h.service.Update(account, data, currentUser)

	if nil != err {
		privateError := errors.PrivateError{Message: "can't update iwt bank account"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("request data", editable)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(updatedIwtBankAccount))
}

// DeleteHandler delete an account
func (h *IwtBankAccountHandler) DeleteHandler(c *gin.Context) {
	logger := h.logger.New("action", "CreateHandler")
	currentUser := h.contextService.MustGetCurrentUser(c)

	id := h.getIdParam(c)
	account, err := h.repo.FindByID(id)

	if nil != err {
		logger.Error("can't retrieve iwt bank account", "err", err, "iwt bank account id", id)
		errcodes.AddError(c, errcodes.CodeIwtBankDetailsNotFound)
		return
	}

	err = h.service.Delete(account, currentUser)
	if nil != err {
		privateError := errors.PrivateError{Message: "can't delete iwt bank account"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("iwt bank account id", id)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusNoContent, response.New())
}

// PdfForAccount generates pdf for account
func (h *IwtBankAccountHandler) PdfForAccount(c *gin.Context) {
	logger := h.logger.New("action", "PdfForAccount")
	currentUser := h.contextService.MustGetCurrentUser(c)

	id := h.getIdParam(c)
	acctID := c.Params.ByName("accountId")
	accID64, err := strconv.ParseUint(acctID, 10, 64)
	if err != nil {
		_ = c.Error(err)
		return
	}

	iwt, err := h.repo.FindByID(id)
	if nil != err {
		logger.Error("can't retrieve iwt bank account", "err", err, "iwt bank account id", id)
		errcodes.AddError(c, errcodes.CodeIwtBankDetailsNotFound)
		return
	}

	acc, err := h.accountRepo.FindByIDAndUserID(accID64, currentUser.UID)
	if nil != err {
		logger.Error("can't retrieve account", "err", err, "account id", accID64)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	b, err := h.pdf.CreatePdfBin(acc, iwt)
	if nil != err {
		_ = c.Error(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=incoming-wire-transfer-%d.pdf", time.Now().Unix()))
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	_, _ = c.Writer.Write(b)
}

// getIdParam returns id or nil
func (h *IwtBankAccountHandler) getIdParam(c *gin.Context) uint64 {
	id := c.Params.ByName("id")

	// convert string to uint
	id64, err := strconv.ParseUint(id, 10, 64)

	if err != nil {
		errcodes.AddError(c, errcodes.CodeNumeric)
		return 0
	}

	return uint64(id64)
}

func (h *IwtBankAccountHandler) getListParams(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, model.IwtBankAccountModel{})
	params.AllowSelectFields(showListOutputFields)
	params.AllowPagination()
	h.addIncludes(params)
	h.addSortings(params)
	return params
}

func (h *IwtBankAccountHandler) addIncludes(params *list_params.ListParams) {
	params.AllowIncludes([]string{"beneficiaryBankDetails", "intermediaryBankDetails", "beneficiaryCustomer"})
}

func (h *IwtBankAccountHandler) addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"beneficiaryBankDetails.bankName", "currencyCode", "isIwtEnabled"})
	params.AddCustomSortings("beneficiaryBankDetails.bankName", h.beneficiaryBankDetailsBankNameSorting)
}

func (h *IwtBankAccountHandler) beneficiaryBankDetailsBankNameSorting(direction string,
	params *list_params.ListParams) (string, error) {
	params.AddLeftJoin("bank_details", "iwt_bank_accounts.beneficiary_bank_details_id = bank_details.id")
	return fmt.Sprintf("bank_details.bank_name %s", direction), nil
}
