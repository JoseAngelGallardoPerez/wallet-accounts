package handlers

import (
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	contextService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/card/service"
)

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
	logger log15.Logger,
) *CsvHandler {
	return &CsvHandler{
		contextService: contextService,
		csvService:     csvService,
		params:         params,
		logger:         logger.New("Handler", "card.CsvHandler"),
	}
}

// AdminsExport handle function for accounts export
func (h *CsvHandler) AdminsExport(c *gin.Context) {
	listParams := h.params.forAdminCsv(c.Request.URL.RawQuery)
	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	file, err := h.csvService.GetCardsFile(listParams)
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
