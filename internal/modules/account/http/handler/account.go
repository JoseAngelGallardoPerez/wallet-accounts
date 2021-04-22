package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/permission"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/account/form"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/account/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/account/service"
	"github.com/Confialink/wallet-accounts/internal/modules/account/wrapper"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	appValidator "github.com/Confialink/wallet-accounts/internal/modules/app/validator"
	authService "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
)

// AccountHandler is handler for accounts requests
type AccountHandler struct {
	repo           *repository.AccountRepository
	service        *service.AccountService
	authService    authService.AuthServiceInterface
	contextService appHttpService.ContextInterface
	csvService     *service.Csv
	validator      appValidator.Interface
	serializer     serializer.AccountSerializerInterface
	creator        *wrapper.AccountCreator
	params         *HandlerParams
	logger         log15.Logger
}

func NewAccountHandler(
	repo *repository.AccountRepository,
	service *service.AccountService,
	authService authService.AuthServiceInterface,
	contextService appHttpService.ContextInterface,
	csvService *service.Csv, validator appValidator.Interface,
	serializer serializer.AccountSerializerInterface,
	creator *wrapper.AccountCreator,
	params *HandlerParams,
	logger log15.Logger,
) *AccountHandler {
	return &AccountHandler{
		repo:           repo,
		service:        service,
		authService:    authService,
		contextService: contextService,
		csvService:     csvService,
		validator:      validator,
		serializer:     serializer,
		creator:        creator,
		params:         params,
		logger:         logger.New("Handler", "AccountHandler"),
	}
}

// ListHandler is handle function for list of accounts
func (self *AccountHandler) ListHandler(c *gin.Context) {
	currentUser := self.contextService.MustGetCurrentUser(c)

	listParams := self.params.forAdmin(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	listParams.AddFilter("userId", []string{currentUser.UID})

	accounts, err := self.service.GetList(listParams)
	if err != nil {
		privateErr := errors.PrivateError{Message: "can't retrieve list of accounts"}
		privateErr.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateErr)
		return
	}

	totalCount, countErr := self.service.GetCount(listParams)
	if countErr != nil {
		privateErr := errors.PrivateError{Message: "can't count accounts"}
		privateErr.AddLogPair("error", countErr.Error())
		errors.AddErrors(c, &privateErr)
		return
	}

	serialized := self.serializer.SerializeList(accounts, listParams.GetOutputFields())
	r := response.NewWithListAndPageLinks(serialized, uint64(totalCount), c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, r)
}

// ListForUserHandler is handle function for list of accounts
func (self *AccountHandler) ListForUserHandler(c *gin.Context) {
	uid := c.Params.ByName("uid")

	listParams := self.params.forUser(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	listParams.AddFilter("userId", []string{uid})
	listParams.AddFilter("isActive", []string{model.AccountStatusIsActiveTrue})

	accounts, err := self.service.GetList(listParams)

	if err != nil {
		privateErr := errors.PrivateError{Message: "can't retrieve list of accounts"}
		privateErr.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateErr)
		return
	}
	serialized := self.serializer.SerializeList(accounts, listParams.GetOutputFields())
	c.JSON(http.StatusOK, serialized)
}

// AdminListHandler returns the list of accounts for admins
func (h *AccountHandler) AdminListHandler(c *gin.Context) {
	listParams := h.params.forAdmin(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	accounts, err := h.service.GetList(listParams)
	if err != nil {
		privateErr := errors.PrivateError{Message: "can't retrieve list of accounts"}
		privateErr.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateErr)
		return
	}
	totalCount, countErr := h.service.GetCount(listParams)
	if countErr != nil {
		privateErr := errors.PrivateError{Message: "can't retrieve list of accounts"}
		privateErr.AddLogPair("error", countErr.Error())
		errors.AddErrors(c, &privateErr)
		return
	}

	serialized := h.serializer.SerializeList(accounts, listParams.GetOutputFields())
	r := response.NewWithListAndPageLinks(serialized, uint64(totalCount), c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, r)
}

// GetHandler returns account by id
func (h *AccountHandler) GetHandler(c *gin.Context) {
	account := h.contextService.GetRequestedAccount(c)
	if account == nil {
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	h.repo.FillUser(account)

	c.JSON(http.StatusOK, response.New().SetData(account))
}

// CreateHandler creates a new account
func (h *AccountHandler) CreateHandler(c *gin.Context) {
	currentUser := h.contextService.MustGetCurrentUser(c)
	action := authService.ActionHas
	resource := authService.ResourcePermission

	var f form.Account
	err := c.ShouldBindJSON(&f)

	if err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	if !f.InitialBalance.IsZero() {
		if !h.authService.CanDynamic(currentUser, action, resource, permission.CreateAccountsWithInitialBalance) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			return
		}
	}

	createdAccount, typedErr := h.creator.CreateAccountWithRequest(&f, h.contextService.MustGetCurrentUser(c))

	if nil != typedErr {
		errors.AddErrors(c, typedErr)
		return
	}

	c.JSON(http.StatusCreated, response.New().SetData(createdAccount))
}

// UpdateHandler updates an account
func (h *AccountHandler) UpdateHandler(c *gin.Context) {
	account := h.contextService.GetRequestedAccount(c)
	if account == nil {
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	currentUser := h.contextService.MustGetCurrentUser(c)

	var editable model.AccountEditable
	err := c.ShouldBindJSON(&editable)
	if err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	err = h.validator.Struct(editable)
	if err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	updatedAccount, pvtErr := h.service.Update(account, &editable, currentUser)
	if nil != pvtErr {
		errors.AddErrors(c, pvtErr)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(updatedAccount))
}

// GenerateNumberHandler returns free account number
func (h *AccountHandler) GenerateNumberHandler(c *gin.Context) {
	f := form.GenerateNumber{}

	_ = c.ShouldBindJSON(&f)

	number := h.service.GenerateAccountNumberWithPrefix(f.Prefix)

	c.JSON(http.StatusOK, response.New().SetData(number))
}

// DeleteHandler delete an account
func (h *AccountHandler) DeleteHandler(c *gin.Context) {
	errcodes.AddError(c, errcodes.CodeForbidden)
}

// ImportCsvHandler imports list of accounts from CSV
func (h *AccountHandler) ImportCsvHandler(c *gin.Context) {
	logger := h.logger.New("action", "ImportCsvHandler")
	currentUser := h.contextService.MustGetCurrentUser(c)

	action := authService.ActionHas
	resource := authService.ResourcePermission

	file, _, err := c.Request.FormFile("file")

	if nil != err {
		logger.Error("can't get file", "err", err)
		errcodes.AddError(c, errcodes.CodeFileInvalid)
		return
	}

	defer file.Close()

	if nil != err {
		pvtErr := errors.PrivateError{Message: err.Error()}
		errors.AddErrors(c, &pvtErr)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		pvtErr := errors.PrivateError{Message: "can't write file to buffer"}
		pvtErr.AddLogPair("error", err.Error())
		errors.AddErrors(c, &pvtErr)
		return
	}

	forms, err := h.csvService.CsvToAccountForms(buf)

	for _, f := range forms {
		if !f.InitialBalance.IsZero() {
			if h.authService.CanDynamic(currentUser, action, resource, permission.CreateAccountsWithInitialBalance) {
				break
			} else {
				errcodes.AddError(c, errcodes.CodeForbidden)
				return
			}
		}
	}

	if nil != err {
		logger.Error("can't convert csv data to accounts structures", "err", err)
		errcodes.AddError(c, errcodes.CodeFileInvalid)
		return
	}

	_, err = h.creator.BulkCreateAccountWithRequest(forms, currentUser)

	if nil != err {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	c.JSON(http.StatusOK, response.New().AddMessage("Accounts successfully imported"))
}
