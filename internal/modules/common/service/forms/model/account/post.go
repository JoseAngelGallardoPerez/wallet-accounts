package account

import (
	"errors"

	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	"github.com/Confialink/wallet-pkg-custom_form"
)

const accountNumberMaxLengthDefault = 28

type Post struct {
	accountTypeRepo *repository.AccountTypeRepository
	defaultMaxValue int
	logger          log15.Logger
}

func NewPost(accountTypeRepo *repository.AccountTypeRepository, logger log15.Logger) *Post {
	return &Post{
		accountTypeRepo,
		accountNumberMaxLengthDefault,
		logger,
	}
}

// make post form
func (s *Post) MakeForm() (*custom_form.Form, error) {
	f := &custom_form.Form{}
	return f, nil
}

// conditions for Number field
func (s *Post) fieldNumberConditions() ([]*custom_form.Condition, error) {
	conditions := make([]*custom_form.Condition, 0, 1)

	manualAccountTypeIds, err := s.manualAccountTypeIds()
	if err != nil {
		return nil, errors.New("cannot receive account types")
	}
	accountTypeCondition := &custom_form.Condition{
		FieldName: "typeId",
		Values:    manualAccountTypeIds,
		Type:      custom_form.ConditionTypeIn,
	}

	conditions = append(conditions, accountTypeCondition)

	return conditions, nil
}

// account types without autogeneration numbers
func (s *Post) manualAccountTypeIds() ([]interface{}, error) {
	params := list_params.NewListParamsFromQuery("", model.AccountType{})
	params.AddFilter("auto_number_generation", []string{"0"}, list_params.OperatorEq)

	manualAccountTypes, err := s.accountTypeRepo.GetList(params)
	if err != nil {
		return nil, errors.New("cannot receive account types")
	}

	manualAccountTypeIds := make([]interface{}, 0, len(manualAccountTypes))
	for _, accountType := range manualAccountTypes {
		manualAccountTypeIds = append(manualAccountTypeIds, accountType.ID)
	}

	return manualAccountTypeIds, nil
}
