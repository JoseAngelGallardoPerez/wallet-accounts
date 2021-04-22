package event

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	"github.com/jinzhu/gorm"
)

const (
	AccountTypeUpdated = "account-type:updated"
)

type ContextAccountTypeUpdated struct {
	// database transaction context
	DbTransaction *gorm.DB
	// model contains data before update
	OldAccountType *model.AccountType
	// mode contains updated data
	NewAccountType *model.AccountType
	// event handler must specify whether processing was finished successfully
	Error error
}
