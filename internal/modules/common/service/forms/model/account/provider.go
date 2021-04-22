package account

import (
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	auth "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
	"github.com/Confialink/wallet-accounts/internal/modules/common/service/forms"
	"github.com/Confialink/wallet-accounts/internal/modules/common/service/forms/interfaces"
)

const ModelName = "account"

// Account form provider
type Provider struct {
	formProviders map[string]interfaces.FormProvider
}

func NewProvider(accountTypeRepo *repository.AccountTypeRepository, logger log15.Logger) interfaces.FormProviderAgregator {
	service := &Provider{}
	postForm := NewPost(accountTypeRepo, logger.New("service", "accountPostForm"))
	service.formProviders = map[string]interfaces.FormProvider{
		forms.TypeFormPost + auth.RoleAdmin: postForm,
		forms.TypeFormPost + auth.RoleRoot:  postForm,
	}

	return service
}

// model name
func (s *Provider) Name() string {
	return ModelName
}

// return registered form providers for the model
func (s *Provider) FormProviders() map[string]interfaces.FormProvider {
	return s.formProviders
}
