package handler

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/account/service"
	contextService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
)

// CsvHandler struct for accounts csv handl functions
type CsvHandler struct {
	contextService contextService.ContextInterface
	csvService     *service.Csv
	params         *HandlerParams
	logger         log15.Logger
}

func NewCsvHandler(
	contextService contextService.ContextInterface,
	csvService *service.Csv,
	params *HandlerParams,
	logger log15.Logger) *CsvHandler {
	return &CsvHandler{contextService: contextService,
		csvService: csvService,
		params:     params,
		logger:     logger.New("Handler", "account.CsvHandler"),
	}
}

// AdminsExport handle function for accounts export
func (h *CsvHandler) AdminsExport(c *gin.Context) {
	listParams := h.params.forAdminCsv(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	file, err := h.csvService.GetAccountsFile(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "Can not get csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	if err := csv.Send(file, c.Writer); err != nil {
		privateError := errors.PrivateError{Message: "Can not send csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
	}
}
