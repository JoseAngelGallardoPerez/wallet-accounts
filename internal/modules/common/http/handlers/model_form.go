package handlers

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/common/service/forms/service"
)

type ModelFormHandler struct {
	modelFormService *service.ModelFormService
	contextService   appHttpService.ContextInterface
	logger           log15.Logger
}

func NewModelFormHandler(
	modelFormService *service.ModelFormService,
	contextService appHttpService.ContextInterface,
	logger log15.Logger,
) *ModelFormHandler {
	return &ModelFormHandler{
		modelFormService,
		contextService,
		logger,
	}
}

func (h *ModelFormHandler) FieldsHandler(c *gin.Context) {
	currentUser := h.contextService.MustGetCurrentUser(c)
	modelName := c.Params.ByName("model")
	formType := c.Params.ByName("type")

	modelForm, typedErr := h.modelFormService.MakeForm(currentUser, modelName, formType)
	if typedErr != nil {
		h.logger.Error(typedErr.Error(), "modelName", modelName, "formType", formType)
		errors.AddErrors(c, typedErr)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(modelForm.Fields))
}
