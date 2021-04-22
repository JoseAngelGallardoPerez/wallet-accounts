package builder

import (
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// DebitFrom declares debit dialog options
type DebitFrom interface {
	// FromAlias specifies existing debitable alias that should be taken debited
	FromAlias(alias string) *Transfer
	// From specifies debitable directly
	From(debitable transfer.Debitable) *Transfer
}

type debitFrom struct {
	transfer            *Transfer
	amount              decimal.Decimal
	currAmount          transfer.CurrencyAmount
	debitActionProvider DebitActionProvider
}

// DebitActionProvider is used in order to
type DebitActionProvider func(transfer.Debitable, transfer.CurrencyAmount) (transfer.Action, error)

// Debit creates new transfer by starting debit dialog
func Debit(amount interface{}) DebitFrom {
	return debit(amount, New())
}

func debit(amount interface{}, t *Transfer) DebitFrom {
	var (
		v                   decimal.Decimal
		currAmount          transfer.CurrencyAmount
		debitActionProvider DebitActionProvider
		err                 error
	)

	switch amount := amount.(type) {
	case string:
		v, err = decimal.NewFromString(amount)
	case int:
		v = decimal.NewFromInt(int64(amount))
	case int64:
		v = decimal.NewFromInt(amount)
	case float32:
		v = decimal.NewFromFloat32(amount)
	case float64:
		v = decimal.NewFromFloat(amount)
	case transfer.CurrencyAmount:
		currAmount = amount
	case decimal.Decimal:
		v = amount
	case *decimal.Decimal:
		v = *amount
	case DebitActionProvider:
		debitActionProvider = amount
	default:
		err = errors.Wrapf(
			transfer.ErrInvalidAmount,
			"failed to create debit action: unexpected amount type is given %T",
			amount,
		)
	}
	if err != nil {
		panic(err)
	}

	return &debitFrom{
		transfer:            t,
		amount:              v,
		currAmount:          currAmount,
		debitActionProvider: debitActionProvider,
	}
}

// From specifies debitable from which to debit
func (d *debitFrom) From(debitable transfer.Debitable) *Transfer {
	var (
		amount      transfer.CurrencyAmount
		debitAction transfer.Action
		err         error
	)
	// currAmount is set when alias is used
	if d.currAmount != nil {
		amount = d.currAmount
	} else {
		amount = transfer.NewAmount(debitable.Currency(), d.amount)
	}
	t := d.transfer

	if d.debitActionProvider != nil {
		debitAction, err = d.debitActionProvider(debitable, amount)
	} else {
		debitAction, err = transfer.NewDebitAction(debitable, amount)
	}

	if err != nil {
		panic(err)
	}
	t.actions = append(t.actions, debitAction)
	t.contextDeb = debitable
	t.contextAction = debitAction
	if purposeSetter, ok := debitAction.(transfer.PurposeSetter); ok {
		t.purposeSetter = purposeSetter
	}
	if messageSetter, ok := debitAction.(transfer.MessageSetter); ok {
		t.messageSetter = messageSetter
	}
	return t
}

// FromAlias specifies debitable alias from which to debit
func (d *debitFrom) FromAlias(alias string) *Transfer {
	t := d.transfer
	debitable, ok := t.debAliases[alias]
	if !ok {
		panic(fmt.Sprintf("mandatory currAmount alias %s is not exist", alias))
	}
	amount := t.mustGetAmount(alias)
	debitAction, err := d.debitActionProvider(debitable, amount)
	if err != nil {
		panic(err)
	}
	t.actions = append(t.actions, debitAction)
	t.contextAction = debitAction
	if purposeSetter, ok := debitAction.(transfer.PurposeSetter); ok {
		t.purposeSetter = purposeSetter
	}
	if messageSetter, ok := debitAction.(transfer.MessageSetter); ok {
		t.messageSetter = messageSetter
	}
	return t
}
