package service

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/common/service/forms/interfaces"
	"github.com/Confialink/wallet-accounts/internal/modules/common/service/forms/model/account"
	"github.com/Confialink/wallet-pkg-custom_form"
	errPkg "github.com/Confialink/wallet-pkg-errors"
	pb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/inconshreveable/log15"
)

type ModelFormService struct {
	forms           map[string]interfaces.FormProviderAgregator
	validatorHelper *custom_form.Helper
}

// here we can register form providers for models
func NewModelFormService(accountTypeRepo *repository.AccountTypeRepository, logger log15.Logger) *ModelFormService {
	service := &ModelFormService{validatorHelper: &custom_form.Helper{}}
	accountFormProvider := account.NewProvider(accountTypeRepo, logger.New("service", "accountFormProvider"))

	// register form providers
	service.forms = map[string]interfaces.FormProviderAgregator{
		accountFormProvider.Name(): accountFormProvider,
	}

	return service
}

// make form for current user and model
func (s *ModelFormService) MakeForm(currentUser *pb.User, modelName, formType string) (*custom_form.Form, errPkg.TypedError) {
	formAgregator, ok := s.forms[modelName]
	if !ok {
		return nil, errcodes.CreatePublicError(errcodes.CodeInvalidFormModel, "invalid form model")
	}

	formProvider, ok := formAgregator.FormProviders()[formType+currentUser.GetRoleName()]
	if !ok {
		return nil, errcodes.CreatePublicError(errcodes.CodeInvalidFormType, "invalid form type")
	}

	form, err := formProvider.MakeForm()
	if err != nil {
		privateError := &errPkg.PrivateError{Message: "cannot make form"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("modelName", modelName)
		privateError.AddLogPair("formType", formType)
		privateError.AddLogPair("role", currentUser.GetRoleName())
		return nil, privateError
	}

	return form, nil
}

// make validator callback to register in struct validator
//func (s *ModelFormService) MakeValidationCallback() validatorPkg.StructLevelFunc {
//	return s.validatorHelper.MakeValidationCallback()
//}
