package errcodes

var details = map[string]string{
	CodeTanEmpty:                        "The request was missing required header X-TAN which must include one time password called TAN",
	CodeTanInvalid:                      "Provided TAN is incorrect or has already been used.",
	CodeTanNotificationMethodNotAllowed: "Notification method is not allowed.",
	CodeTanRequestNotAllowed:            "TAN request is not allowed.",
	CodeAccountNotFound:                 "Account was not found.",
	CodeUserNotFound:                    "User was not found",
	CodeDuplicateTransferTemplate:       "Template with the same name and request subject is already exist.",
	CodeWithdrawalNotAllowed:            "It is not allowed to perform withdrawals from this account.",
	CodeDepositNotAllowed:               "It is not allowed to perform deposit to this account.",
	CodeInsufficientFunds:               "You don't have enough funds on this account. Reduce the amount or replenish the balance.",
	CodeForbidden:                       "You are not allowed to perform this action.",
	CodeDuplicateTransferFee:            "Transfer fee with the same name and request subject is already exist.",
	CodeUnknownRequestSubject:           "Unknown request subject.",
	CodeAccountInactive:                 "Account is not active.",
	CodeLimitExceeded:                   "The requested action could not be performed due to the limitations that will be exceeded as a result of this action.",
	CodeExchangeRateNotFound:            "The requested action requires a currency exchange rate that is currently not available.",
}
