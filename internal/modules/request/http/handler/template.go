package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-response"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	httpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	authService "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	templateRepository "github.com/Confialink/wallet-accounts/internal/modules/request/repository"
)

type TemplateHandler struct {
	authService        *authService.AuthService
	contextService     httpService.ContextInterface
	templateRepository *templateRepository.Template
	logger             log15.Logger
}

func NewTemplateHandler(
	contextService httpService.ContextInterface,
	templateRepository *templateRepository.Template,
	logger log15.Logger,
) *TemplateHandler {
	return &TemplateHandler{
		contextService:     contextService,
		templateRepository: templateRepository,
		logger:             logger.New("Handler", "TemplateHandler"),
	}
}

func (t *TemplateHandler) List(c *gin.Context) {
	user := t.contextService.MustGetCurrentUser(c)

	requestSubject := c.Param("subject")

	subject, publicErr := constants.SubjectFromString(requestSubject)
	if publicErr != nil {
		errors.AddErrors(c, publicErr)
		return
	}

	templates, err := t.templateRepository.FindByUserIdAndRequestSubject(user.UID, subject)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeTemplateNotFound)
		return
	}

	c.JSON(http.StatusOK, response.NewResponse(templates))
}

func (t *TemplateHandler) ListAll(c *gin.Context) {
	user := t.contextService.MustGetCurrentUser(c)

	templates, err := t.templateRepository.FindByUserId(user.UID)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeTemplateNotFound)
		return
	}

	c.JSON(http.StatusOK, response.NewResponse(templates))
}

func (t *TemplateHandler) Delete(c *gin.Context) {
	user := t.contextService.MustGetCurrentUser(c)

	logger := t.logger.New("action", "Delete")
	id, typedErr := t.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	template, err := t.templateRepository.FindById(id)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			errcodes.AddError(c, errcodes.CodeTemplateNotFound)
			return
		}
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		logger.Error("failed to retrieve template", "error", err, "id", id)
		return
	}

	if *template.UserId != user.UID {
		errcodes.AddError(c, errcodes.CodeForbidden)
		return
	}

	err = t.templateRepository.Delete(template)
	if err != nil {
		privateError := errors.PrivateError{Message: "unable to delete record"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	c.Status(http.StatusOK)
}
