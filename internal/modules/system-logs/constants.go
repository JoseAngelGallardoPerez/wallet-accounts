package system_logs

const (
	SubjectManualTransaction = "Manual transaction"
	SubjectRevenueDeduction  = "Revenue deduction"

	SubjectCreateIwtBankAccounts = "Create IWT Bank Accounts"
	SubjectDeleteIwtBankAccounts = "Delete IWT Bank Accounts"
	SubjectModifyIwtBankAccounts = "Modify IWT Bank Accounts"

	SubjectCreateAccount = "New Account"
	SubjectModifyAccount = "Modify Account"

	SubjectCreateAccountTypes = "New Account Type"
	SubjectModifyAccountTypes = "Modify Account Type"

	SubjectCreateCard = "New Card"
	SubjectModifyCard = "Modify Card"

	SubjectCreateCardType = "New Card Type"
	SubjectModifyCardType = "Modify Card Type"
)

const (
	DataTitleMessage                     = "Message"
	DataTitleBankAccountReference        = "Bank Account Reference"
	DataTitleBankAccountReferenceDetails = "Bank Account Reference Details"
	DataTitleAccountDetails              = "Account Details"
	DataTitleAccountTypeDetails          = "Account Type Details"
	DataTitleCardDetails                 = "Card Details"
	DataTitleCardTypeDetails             = "Card Type Details"
)

const (
	OperationTypeCredit = "credit"
	OperationTypeDebit  = "debit"
)
