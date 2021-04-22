package initializers

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	appValidator "github.com/Confialink/wallet-accounts/internal/modules/app/validator"
	cardTypeFormat "github.com/Confialink/wallet-accounts/internal/modules/card-type-format/model"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/repository"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	cardRepository "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	currencyService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/shopspring/decimal"

	"github.com/go-playground/validator/v10"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"reflect"
	"regexp"
)

const (
	alphaNumericRegexString = "^[a-zA-Z0-9]+$"
	numericRegexString      = "^[-+]?[0-9]+(?:\\.[0-9]+)?$"
	decimalRegexString      = "^-?\\d+(\\.\\d+)?$"

	tagExistUserId         = "existUserId"
	tagUserIsActive        = "userIsActive"
	tagAccountNumberUnique = "accountNumberUnique"
	tagCardNumberUnique    = "cardNumberUnique"
)

var (
	usersService      *service.UserService
	currenciesService currencyService.CurrenciesServiceInterface
	cardTypeRepo      repository.CardTypeRepositoryInterface
	cardRepo          cardRepository.CardRepositoryInterface
	accountRepo       *accountRepository.AccountRepository
	logger            log15.Logger

	alphaNumericRegex = regexp.MustCompile(alphaNumericRegexString)
	numericRegex      = regexp.MustCompile(numericRegexString)
	decimalRegex      = regexp.MustCompile(decimalRegexString)
)

func LoadDependencies(
	usersServiceDep *service.UserService,
	currenciesServiceDep currencyService.CurrenciesServiceInterface,
	cardTypeRepoDep repository.CardTypeRepositoryInterface,
	cardRepoDep cardRepository.CardRepositoryInterface,
	accountRepoDep *accountRepository.AccountRepository,
	loggerDep log15.Logger,
) {
	usersService = usersServiceDep
	currenciesService = currenciesServiceDep
	cardTypeRepo = cardTypeRepoDep
	cardRepo = cardRepoDep
	accountRepo = accountRepoDep
	logger = loggerDep.New("where", "config.initializers.Validator")
}

// Initialize executes code should be initialized before server starting
func Initialize(validator appValidator.Interface) {
	_ = validator.RegisterValidation("decimal", decimalValid)
	_ = validator.RegisterValidation("decimalGT", decimalGreaterThan)
	_ = validator.RegisterValidation(tagExistUserId, existingUserIDValidation)
	_ = validator.RegisterValidation(tagUserIsActive, userIsActiveValidation)
	_ = validator.RegisterValidation("activeCurrencyCode", activeCurrencyCodeValidation)
	_ = validator.RegisterValidation("validCardFormat", validCardFormat)
	_ = validator.RegisterValidation("notRequiredNumericPointer", notRequiredNumericPointer)
	_ = validator.RegisterValidation("notRequiredAlphanumericPointer", notRequiredAlphanumericPointer)
	_ = validator.RegisterValidation(tagAccountNumberUnique, accountNumberUnique)
	_ = validator.RegisterValidation(tagCardNumberUnique, cardNumberUnique)

	registerFormatters()
}

// register formatters for custom tags
func registerFormatters() {
	formatters := map[string]*errors.ValidationErrorFormatter{
		tagAccountNumberUnique: {
			Code:      errcodes.CodeDuplicateAccountNumber,
			TitleFunc: func(_ validator.FieldError, formattedField string) string { return "" },
		},
		tagCardNumberUnique: {
			Code:      errcodes.CodeDuplicateCardNumber,
			TitleFunc: func(_ validator.FieldError, formattedField string) string { return "" },
		},
		tagExistUserId: {
			Code:      errcodes.CodeUserNotFound,
			TitleFunc: func(_ validator.FieldError, formattedField string) string { return "" },
		},
		tagUserIsActive: {
			Code:      errcodes.CodeUserMustBeActive,
			TitleFunc: func(_ validator.FieldError, formattedField string) string { return "" },
		},
	}
	errors.SetFormatters(formatters)
}

func existingUserIDValidation(fl validator.FieldLevel) bool {
	fieldStr := fl.Field().Interface().(string)
	_, err := usersService.GetByUID(fieldStr)
	return err == nil
}

// valid if a user is active
func userIsActiveValidation(fl validator.FieldLevel) bool {
	fieldStr := fl.Field().Interface().(string)

	users, err := usersService.GetFullByUIDs([]string{fieldStr}, []string{"Status"})
	if err != nil || len(users) == 0 {
		return false
	}

	return users[0].Status == "active"
}

func activeCurrencyCodeValidation(fl validator.FieldLevel) bool {
	fieldStr := fl.Field().Interface().(string)
	currency, err := currenciesService.GetByCode(fieldStr)
	if err != nil {
		logger.Error("Filed to get currency for validation", "error", err)
		return false
	}
	return currency.Active
}

func validCardFormat(fl validator.FieldLevel) bool {
	fieldStr := fl.Field().Interface().(string)
	card := fl.Parent().Elem().Interface().(cardModel.Card)
	includes := list_params.Includes{}
	includes.AddIncludes("format")

	cardType, err := cardTypeRepo.Get(*card.CardTypeId, &includes)
	if err != nil {
		return false
	}

	if cardType.Format == nil {
		return false
	}

	if pattern, ok := cardTypeFormat.Rules[*cardType.Format.Code]; ok {
		compiledPattern := regexp.MustCompile(pattern)
		return compiledPattern.MatchString(fieldStr)
	}

	return false
}

func notRequiredNumericPointer(fl validator.FieldLevel) bool {
	fieldStr := fl.Field().Interface().(string)

	if len(fieldStr) == 0 {
		return true
	}

	return isNumeric(fl)
}

func notRequiredAlphanumericPointer(fl validator.FieldLevel) bool {
	fieldStr := fl.Field().Interface().(string)

	if len(fieldStr) == 0 {
		return true
	}

	return isAlphanum(fl)
}

func accountNumberUnique(fl validator.FieldLevel) bool {
	number := fl.Field().Interface().(string)
	acc, err := accountRepo.FindByNumber(number)
	if (err != nil && err != gorm.ErrRecordNotFound) || acc != nil {
		return false
	}

	return true
}

func cardNumberUnique(fl validator.FieldLevel) bool {
	number := fl.Field().Interface().(string)
	card, err := cardRepo.GetByNumber(number, nil)
	if (err != nil && err != gorm.ErrRecordNotFound) || card != nil {
		return false
	}

	return true
}

// Valid checks if passed value is valid decimal number
func decimalValid(fl validator.FieldLevel) bool {
	field := fl.Field()
	fieldKind := field.Kind()
	if fieldKind != reflect.String {
		return false
	}
	fieldStr := field.Interface().(string)
	return decimalRegex.MatchString(fieldStr)
}

// GreaterThan checks if passed decimal value is greater than specified.
// Validator usage: "YOUR_TAG=DECIMAL_NUMBER", e.g. "decimalGT=0"
func decimalGreaterThan(fl validator.FieldLevel) bool {
	field := fl.Field()
	param := fl.Param()
	fieldStr := field.Interface().(string)
	dec, err := decimal.NewFromString(fieldStr)
	if err != nil {
		return false
	}
	decParam, _ := decimal.NewFromString(param)
	return dec.GreaterThan(decParam)
}

// IsNumeric is the validation function for validating if the current field's value is a valid numeric value.
func isNumeric(fl validator.FieldLevel) bool {
	switch fl.Field().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
		return true
	default:
		return numericRegex.MatchString(fl.Field().String())
	}
}

// IsAlphanum is the validation function for validating if the current field's value is a valid alphanumeric value.
func isAlphanum(fl validator.FieldLevel) bool {
	return alphaNumericRegex.MatchString(fl.Field().String())
}
