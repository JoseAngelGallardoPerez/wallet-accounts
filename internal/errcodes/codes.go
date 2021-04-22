package errcodes

import "net/http"

const (
	CodeTanEmpty                        = "TAN_EMPTY"
	CodeTanInvalid                      = "TAN_INVALID"
	CodeTanNotificationMethodNotAllowed = "TAN_NOTIFICATION_METHOD_NOT_ALLOWED"
	CodeTanRequestNotAllowed            = "TAN_REQUEST_NOT_ALLOWED"
	CodeAccountNotFound                 = "ACCOUNT_NOT_FOUND"
	CodeUserNotFound                    = "USER_NOT_FOUND"
	CodeDuplicateTransferTemplate       = "DUPLICATE_TRANSFER_TEMPLATE"
	CodeDuplicateTransferFee            = "DUPLICATE_TRANSFER_FEE"
	CodeUncoveredTransferFee            = "UNCOVERED_TRANSFER_FEE"
	CodeWithdrawalNotAllowed            = "WITHDRAWAL_NOT_ALLOWED"
	CodeDepositNotAllowed               = "DEPOSIT_NOT_ALLOWED"
	CodeAccountInactive                 = "ACCOUNT_INACTIVE"
	CodeInsufficientFunds               = "INSUFFICIENT_FUNDS"
	CodeLimitExceeded                   = "LIMIT_EXCEEDED"
	CodeForbidden                       = "FORBIDDEN"
	CodeFeeParamsNotFound               = "FEE_PARAMS_NOT_FOUND"
	CodeFileInvalid                     = "FILE_INVALID"
	CodeFileEmpty                       = "FILE_EMPTY"
	CodeRequestNotFound                 = "REQUEST_NOT_FOUND"
	CodeUnknownRequestStatus            = "UNKNOWN_REQUEST_STATUS"
	CodeNumeric                         = "NUMERIC"
	CodeUnknownRequestSubject           = "UNKNOWN_REQUEST_SUBJECT"
	CodeAccountTypeContainsAccounts     = "ACCOUNT_TYPE_CONTAINS_ACCOUNTS"
	CodeTransactionNotFound             = "TRANSACTION_NOT_FOUND"
	CodeDuplicateAccountNumber          = "DUPLICATE_ACCOUNT_NUMBER"
	CodeAccountTypeNotFound             = "ACCOUNT_TYPE_NOT_FOUND"
	CodeInvalidAccountOwner             = "INVALID_ACCOUNT_OWNER"
	CodeRatesDoNotMatch                 = "RATES_DO_NOT_MATCH"
	CodeInvalidExchangeRate             = "INVALID_EXCHANGE_RATE"
	CodeExchangeRateNotFound            = "EXCHANGE_RATE_NOT_FOUND"
	CodeTemplateNotFound                = "TEMPLATE_NOT_FOUND"
	CodeCardNotFound                    = "CARD_NOT_FOUND"
	CodeDuplicateCardNumber             = "DUPLICATE_CARD_NUMBER"
	CodeInvalidCardOwner                = "INVALID_CARD_OWNER"
	CodeInvalidTemplate                 = "INVALID_TEMPLATE"
	CodeCardTypeCategoryNotFound        = "CARD_TYPE_CATEGORY_NOT_FOUND"
	CodeCardTypeFormatNotFound          = "CARD_TYPE_FORMAT_NOT_FOUND"
	CodeInvalidQueryParameters          = "INVALID_QUERY_PARAMETERS"
	CodeAmountsDoNatMatch               = "AMOUNTS_DO_NOT_MATCH"
	CodeSettingNotFound                 = "SETTING_NOT_FOUND"
	CodeIwtBankDetailsNotFound          = "IWT_BANK_DETAILS_NOT_FOUND"
	CodeCardTypeNotFound                = "CARD_TYPE_NOT_FOUND"
	CodeCardTypeAssociatedWithCards     = "CARD_TYPE_ASSOCIATED_WITH_CARDS"
	CodeCsvFileInvalidRow               = "CSV_FILE_INVALID_ROW"
	CodeUserInvalidStatus               = "USER_INVALID_STATUS"
	CodeUserMustBeActive                = "USER_MUST_HAVE_ACTIVE_STATUS"
	CardTypeNameIsDuplicated            = "CARD_TYPE_NAME_IS_DUPLICATED"
	AccountTypeNameIsDuplicated         = "ACCOUNT_TYPE_NAME_IS_DUPLICATED"
	CodeCurrencyMismatch                = "CURRENCY_MISMATCH"
	CodeInvalidCurrencyPrecision        = "INVALID_CURRENCY_PRECISION"

	CodeInvalidFormModel = "INVALID_FORM_MODEL"
	CodeInvalidFormType  = "INVALID_FORM_TYPE"
)

func HttpStatusCodeByErrCode(code string) int {
	if status, ok := statusCodes[code]; ok {
		return status
	}
	panic("code is not present")
}

func IsKnownCode(code string) bool {
	_, ok := statusCodes[code]
	return ok
}

var statusCodes = map[string]int{
	CodeTanEmpty:                        http.StatusBadRequest,
	CodeTanInvalid:                      http.StatusForbidden,
	CodeTanNotificationMethodNotAllowed: http.StatusForbidden,
	CodeTanRequestNotAllowed:            http.StatusUnprocessableEntity,
	CodeAccountNotFound:                 http.StatusNotFound,
	CodeUserNotFound:                    http.StatusNotFound,
	CodeDuplicateTransferTemplate:       http.StatusConflict,
	CodeWithdrawalNotAllowed:            http.StatusUnprocessableEntity,
	CodeDepositNotAllowed:               http.StatusUnprocessableEntity,
	CodeInsufficientFunds:               http.StatusUnprocessableEntity,
	CodeLimitExceeded:                   http.StatusUnprocessableEntity,
	CodeAccountInactive:                 http.StatusUnprocessableEntity,
	CodeUncoveredTransferFee:            http.StatusConflict,
	CodeForbidden:                       http.StatusForbidden,
	CodeDuplicateTransferFee:            http.StatusConflict,
	CodeFeeParamsNotFound:               http.StatusNotFound,
	CodeFileInvalid:                     http.StatusBadRequest,
	CodeUnknownRequestStatus:            http.StatusBadRequest,
	CodeNumeric:                         http.StatusBadRequest,
	CodeRequestNotFound:                 http.StatusNotFound,
	CodeFileEmpty:                       http.StatusBadRequest,
	CodeUnknownRequestSubject:           http.StatusBadRequest,
	CodeAccountTypeContainsAccounts:     http.StatusBadRequest,
	CodeTransactionNotFound:             http.StatusNotFound,
	CodeDuplicateAccountNumber:          http.StatusConflict,
	CodeAccountTypeNotFound:             http.StatusNotFound,
	CodeInvalidAccountOwner:             http.StatusForbidden,
	CodeRatesDoNotMatch:                 http.StatusBadRequest,
	CodeInvalidExchangeRate:             http.StatusBadRequest,
	CodeExchangeRateNotFound:            http.StatusNotFound,
	CodeTemplateNotFound:                http.StatusNotFound,
	CodeCardNotFound:                    http.StatusNotFound,
	CodeInvalidCardOwner:                http.StatusBadRequest,
	CodeInvalidTemplate:                 http.StatusBadRequest,
	CodeCardTypeCategoryNotFound:        http.StatusNotFound,
	CodeCardTypeFormatNotFound:          http.StatusNotFound,
	CodeInvalidQueryParameters:          http.StatusBadRequest,
	CodeAmountsDoNatMatch:               http.StatusBadRequest,
	CodeSettingNotFound:                 http.StatusNotFound,
	CodeIwtBankDetailsNotFound:          http.StatusNotFound,
	CodeCardTypeNotFound:                http.StatusNotFound,
	CodeCardTypeAssociatedWithCards:     http.StatusConflict,
	CodeCsvFileInvalidRow:               http.StatusBadRequest,
	CodeUserInvalidStatus:               http.StatusBadRequest,
	CodeUserMustBeActive:                http.StatusBadRequest,
	CodeDuplicateCardNumber:             http.StatusBadRequest,
	CodeInvalidFormModel:                http.StatusBadRequest,
	CodeInvalidFormType:                 http.StatusBadRequest,
	CardTypeNameIsDuplicated:            http.StatusBadRequest,
	AccountTypeNameIsDuplicated:         http.StatusBadRequest,
}
