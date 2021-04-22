package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
	"github.com/Confialink/wallet-accounts/internal/modules/settings/form"
	"github.com/Confialink/wallet-accounts/internal/modules/settings/repository"
)

type SettingsController struct {
	repository     *repository.Settings
	service        *settings.Service
	contextService appHttpService.ContextInterface
}

func NewSettingsController(
	repository *repository.Settings,
	service *settings.Service,
	contextService appHttpService.ContextInterface,
) *SettingsController {
	return &SettingsController{
		repository:     repository,
		service:        service,
		contextService: contextService,
	}
}

func (s *SettingsController) UpdateSetting(c *gin.Context) {
	settingForm := &struct {
		Value string `json:"value" settingForm:"value" binding:"required"`
	}{}

	if err := c.ShouldBind(settingForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	settingName := c.Param("setting")

	setting, err := s.repository.Update(settingName, settingForm.Value)
	if nil != err {
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(setting))
}

func (s *SettingsController) MassUpdateSetting(c *gin.Context) {
	settingsForm := &struct {
		KeyValuePairs []*form.KeyValue `json:"keyValuePairs" binding:"required,dive"`
	}{}

	if err := c.ShouldBind(settingsForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	err := s.repository.MassUpdate(settingsForm.KeyValuePairs)
	if nil != err {
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (s *SettingsController) ListSettings(c *gin.Context) {
	settings, err := s.repository.GetAll()
	if nil != err {
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		return
	}

	list := response.NewList(settings)

	c.JSON(http.StatusOK, response.New().SetData(list.Items))
}

func (s *SettingsController) GetSetting(c *gin.Context) {
	settingName := c.Param("setting")

	setting, err := s.repository.FirstByName(settingName)
	if nil != err {
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		return
	}

	if !setting.IsExist() {
		errcodes.AddError(c, errcodes.CodeSettingNotFound)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(setting))
}
