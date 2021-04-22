package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Executor is used in order to "execute" request
type Executor interface {
	Execute(request *model.Request) (types.Details, error)
}

// CreateExecutor is a factory func that provides Executor based on request subject
func CreateExecutor(
	db *gorm.DB, request *model.Request,
	provider transfer.CurrencyProvider,
	pf PermissionFactory,
) (Executor, error) {
	switch request.Subject.String() {
	case "TBA", "TBU":
		return baTransfer(db, request, provider, pf), nil
	case "OWT":
		return owTransfer(db, request, provider, pf), nil
	case "CFT":
		return cfTransfer(db, request, provider, pf), nil

	}
	return nil, errors.Wrapf(
		ErrSubjectNotSupported,
		`executor cannot be created, subject "%s" is not supported`,
		request.Subject.String(),
	)
}

// Canceller is used in order to "cancel" pending request
type Canceller interface {
	Cancel(request *model.Request, reason string) error
}

// CreateCanceller is a factory func that provides Canceller based on request subject
func CreateCanceller(
	db *gorm.DB, request *model.Request,
	provider transfer.CurrencyProvider,
	pf PermissionFactory,
) (Canceller, error) {
	switch request.Subject.String() {
	case "TBA", "TBU":
		return baTransfer(db, request, provider, pf), nil
	case "OWT":
		return owTransfer(db, request, provider, pf), nil
	case "CFT":
		return cfTransfer(db, request, provider, pf), nil
	}
	return nil, errors.Wrapf(
		ErrSubjectNotSupported,
		`canceller cannot be created, subject "%s" is not supported`,
		request.Subject.String(),
	)
}

// Modifier is used in order to "modify" pending request
type Modifier interface {
	Modify(request *model.Request) (types.Details, error)
}

// CreateCanceller is a factory func that provides Modifier based on request subject
func CreateModifier(
	db *gorm.DB, request *model.Request,
	provider transfer.CurrencyProvider,
	pf PermissionFactory,
) (Modifier, error) {
	switch request.Subject.String() {
	case "TBA", "TBU":
		return baTransfer(db, request, provider, pf), nil
	case "OWT":
		return owTransfer(db, request, provider, pf), nil
	case "CFT":
		return cfTransfer(db, request, provider, pf), nil
	}
	return nil, errors.Wrapf(
		ErrSubjectNotSupported,
		`modifier cannot be created, subject "%s" is not supported`,
		request.Subject.String(),
	)
}
