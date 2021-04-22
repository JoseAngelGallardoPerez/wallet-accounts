package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-response"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
)

type CsvHandler struct {
	contextService     service.ContextInterface
	csvService         *request.CsvService
	params             *HandlerPrams
	logger             log15.Logger
	accountsRepository *repository.AccountRepository
}

func NewCsvHandler(
	contextService service.ContextInterface,
	csvService *request.CsvService,
	params *HandlerPrams,
	logger log15.Logger,
	accountsRepository *repository.AccountRepository,
) *CsvHandler {
	return &CsvHandler{
		contextService:     contextService,
		csvService:         csvService,
		params:             params,
		logger:             logger.New("Handler", "request.CsvHandler"),
		accountsRepository: accountsRepository,
	}
}

// UpdateFromCsv updates list of requests from csv
// File format: request id, status, rate
func (r *CsvHandler) UpdateFromCsv(c *gin.Context) {
	logger := r.logger.New("action", "UpdateFromCsv")

	file, _, err := c.Request.FormFile("file")

	if nil != err {
		logger.Error("can't get file", "err", err)
		errcodes.AddError(c, errcodes.CodeFileInvalid)
		return
	}

	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		privateError := errors.PrivateError{Message: "can't write file to buffer"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	user := r.contextService.MustGetCurrentUser(c)
	tErrs, failed, success := r.csvService.UpdateFromCsv(buf, user)
	data := struct {
		Success uint64 `json:"success"`
		Failed  uint64 `json:"failed"`
	}{
		Success: success,
		Failed:  failed,
	}

	if len(tErrs) > 0 {
		c.Set("data", data)
		errors.AddErrors(c, tErrs...)
		return
	}

	res := response.NewResponse().AddMessage("Requests are successfully updated")
	res.Data = data
	c.JSON(http.StatusOK, res)
}

// ImportFromCsv imports list of requests from csv
// File format: account number, debit or credit, amount, description, revenue, apply IWT fee
func (r *CsvHandler) ImportFromCsv(c *gin.Context) {
	logger := r.logger.New("action", "UpdateFromCsv")
	user := r.contextService.MustGetCurrentUser(c)

	file, _, err := c.Request.FormFile("file")

	if nil != err {
		logger.Error("can't get file", "err", err)
		errcodes.AddError(c, errcodes.CodeFileInvalid)
		return
	}

	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		privateError := errors.PrivateError{Message: "can't write file to buffer"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	tErrs, failed, success := r.csvService.ImportFromCsv(buf, user)
	data := struct {
		Success uint64 `json:"success"`
		Failed  uint64 `json:"failed"`
	}{
		Success: success,
		Failed:  failed,
	}

	if len(tErrs) > 0 {
		c.Set("data", data)
		errors.AddErrors(c, tErrs...)
		return
	}

	res := response.NewResponse().AddMessage("Requests are successfully imported")
	res.Data = data
	c.JSON(http.StatusOK, res)
}

func (h *CsvHandler) DownloadAdminsReport(c *gin.Context) {
	user := h.contextService.MustGetCurrentUser(c)

	listParams := h.params.adminCsv(c.Request.URL.RawQuery)
	// TODO: refactor is_visible field is actually used in order to filter requests which are initiated by admin
	// or has been approved without confirmation, we should remove is_visible field and add instantly_approved field instead
	listParams.AddFilter("isVisible", []string{"true"})
	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	file, err := h.csvService.GetCsvFile(listParams, user.RoleName)
	if err != nil {
		privateError := errors.PrivateError{Message: "Can not get csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	if err = csv.Send(file, c.Writer); err != nil {
		privateError := errors.PrivateError{Message: "Can not send csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
	}
}
