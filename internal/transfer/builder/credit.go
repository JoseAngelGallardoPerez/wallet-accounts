package builder

import (
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// CreditTo declares credit dialog options
type CreditTo interface {
	// ToCurrency specifies destination creditable instance
	To(creditable transfer.Creditable) *Transfer
}

type creditTo struct {
	transfer   *Transfer
	currAmount transfer.CurrencyAmount
	amount     decimal.Decimal
}

// Credit creates new transfer by starting credit dialog
func Credit(amount interface{}) CreditTo {
	return credit(amount, New())
}

func credit(amount interface{}, t *Transfer) CreditTo {
	var (
		v          decimal.Decimal
		currAmount transfer.CurrencyAmount
		err        error
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
	default:
		err = errors.Wrapf(
			transfer.ErrInvalidAmount,
			"failed to create credit action: unexpected amount type is given %T",
			amount,
		)
	}
	if err != nil {
		panic(err)
	}

	return &creditTo{
		transfer:   t,
		amount:     v,
		currAmount: currAmount,
	}
}

// ToCurrency specifies destination creditable instance
func (c *creditTo) To(creditable transfer.Creditable) *Transfer {
	var (
		amount       transfer.CurrencyAmount
		creditAction transfer.Action
		err          error
	)
	// currAmount is set when alias is used
	if c.currAmount != nil {
		amount = c.currAmount
	} else {
		amount = transfer.NewAmount(creditable.Currency(), c.amount)
	}

	creditAction, err = transfer.NewCreditAction(creditable, amount)

	if err != nil {
		panic(err)
	}
	c.transfer.actions = append(c.transfer.actions, creditAction)

	if purposeSetter, ok := creditAction.(transfer.PurposeSetter); ok {
		c.transfer.purposeSetter = purposeSetter
	}
	if messageSetter, ok := creditAction.(transfer.MessageSetter); ok {
		c.transfer.messageSetter = messageSetter
	}

	c.transfer.contextAction = creditAction
	c.transfer.contextCred = creditable

	return c.transfer
}
