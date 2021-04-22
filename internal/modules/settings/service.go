package settings

import (
	"log"
	"strconv"

	"github.com/Confialink/wallet-accounts/internal/modules/settings/model"
	"github.com/Confialink/wallet-accounts/internal/modules/settings/repository"
)

type Service struct {
	repository *repository.Settings
}

func NewService(repository *repository.Settings) *Service {
	return &Service{repository: repository}
}

//Bool retrieves setting by its name and interpolate string value. "true" or "1" are the only strings treated as boolean true
func (s *Service) Bool(name Name) (bool, error) {
	setting, err := s.fetchSetting(name)
	if nil != err {
		return false, err
	}

	return setting.Value == "true" || setting.Value == "1", nil
}

func (s *Service) String(name Name) (string, error) {
	setting, err := s.fetchSetting(name)
	if nil != err {
		return "", err
	}
	return setting.Value, nil
}

func (s *Service) Int64(name Name) (int64, error) {
	setting, err := s.fetchSetting(name)
	if nil != err {
		return 0, err
	}
	result, err := strconv.ParseInt(setting.Value, 10, 64)
	if nil != err {
		return 0, err
	}
	return result, nil
}

func (s *Service) fetchSetting(name Name) (*model.Settings, error) {
	setting, err := s.repository.FirstByName(name.String())
	if nil != err {
		log.Printf("Settings service: error while fetching setting %s - %s", name, err.Error())
		return nil, err
	}
	return setting, nil
}
