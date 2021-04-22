package builder

import (
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"fmt"
)

// Transfer is used in order to compose and execute a set of actions
type Transfer struct {
	contextDeb    transfer.Debitable
	contextCred   transfer.Creditable
	contextAction transfer.Action
	debAliases    map[string]transfer.Debitable
	actionAliases map[string]transfer.Action
	amountAliases map[string]transfer.CurrencyAmount
	actions       []transfer.Action
	groups        map[string]Group

	purposeSetter transfer.PurposeSetter
	messageSetter transfer.MessageSetter
}

func New() *Transfer {
	return &Transfer{
		debAliases:    make(map[string]transfer.Debitable),
		amountAliases: make(map[string]transfer.CurrencyAmount),
		actionAliases: make(map[string]transfer.Action),
		actions:       make([]transfer.Action, 0, 4),
		groups:        make(map[string]Group),
	}
}

// Execute performs a set of actions
func (t *Transfer) Execute() error {
	for _, action := range t.actions {
		if err := action.Perform(); err != nil {
			return err
		}
	}
	return nil
}

// Actions returns list of transfer actions
func (t *Transfer) Actions() []transfer.Action {
	return t.actions
}

// AmountAlias retrieves alias amount by name
func (t *Transfer) AmountAlias(alias string) transfer.CurrencyAmount {
	for name, amount := range t.amountAliases {
		if name == alias {
			return amount
		}
	}
	return nil
}

// WithPurpose sets purpose using current context
func (t *Transfer) WithPurpose(purpose string) *Transfer {
	if t.purposeSetter != nil {
		t.purposeSetter.SetPurpose(purpose)
	}
	return t
}

// WithMessage sets message using current context
func (t *Transfer) WithMessage(message string) *Transfer {
	if t.messageSetter != nil {
		t.messageSetter.SetMessage(message)
	}
	return t
}

// Debit opens debit action creation dialog
// amount could be string, int, float, decimal.Decimal, transfer.CurrencyAmount, DebitActionProvider
func (t *Transfer) Debit(amount interface{}) DebitFrom {
	return debit(amount, t)
}

// DebitFromAlias opens debit action creation dialog
// alias should specify existing amount alias
func (t *Transfer) DebitFromAlias(alias string) DebitFrom {
	return debit(t.mustGetAmount(alias), t)
}

// Credit opens credit action creation dialog
// amount could be string, int, float, decimal.Decimal, transfer.CurrencyAmount
func (t *Transfer) Credit(amount interface{}) CreditTo {
	return credit(amount, t)
}

// CreditFromAlias opens credit action creation dialog
// alias should specify existing amount alias
func (t *Transfer) CreditFromAlias(alias string) CreditTo {
	return credit(t.mustGetAmount(alias), t)
}

// Exchange opens exchange action creation dialog
func (t *Transfer) Exchange(from transfer.CurrencyAmount) ExchangeUsing {
	return exchangeU(from, t)
}

// ExchangeFromAlias opens exchange action creation dialog
// alias should specify existing amount alias
func (t *Transfer) ExchangeFromAlias(alias string) ExchangeUsing {
	return &exchangeUsing{
		transfer:       t,
		fromCurrAmount: t.mustGetAmount(alias),
	}
}
func (t *Transfer) WithCallback(callback func(action transfer.Action) error) *Transfer {
	if t.contextAction != nil {
		t.contextAction = transfer.NewHookAction(t.contextAction, callback)
		t.actions[len(t.actions)-1] = t.contextAction
	}
	return t
}

// As creates new alias
func (t *Transfer) As(alias string) *Transfer {
	if t.contextDeb != nil {
		t.debAliases[alias] = t.contextDeb
		t.amountAliases[alias] = t.contextDeb
	}
	if t.contextAction != nil {
		t.actionAliases[alias] = t.contextAction
		t.amountAliases[alias] = t.contextAction
	}
	return t
}

// IncludeToGroup adds action amount to a named group
func (t *Transfer) IncludeToGroup(name string) *Transfer {
	if t.contextAction != nil {
		group, ok := t.groups[name]
		if !ok {
			t.groups[name] = Group{t.contextAction}
			return t
		}
		t.groups[name] = append(group, t.contextAction)
	}
	return t
}

// GetGroup returns group by a given name
func (t *Transfer) GetGroup(name string) Group {
	return t.groups[name]
}

func (t *Transfer) mustGetAmount(alias string) transfer.CurrencyAmount {
	amount, ok := t.actionAliases[alias]
	if !ok {
		panic(fmt.Sprintf("mandatory currAmount alias %s is not exist", alias))
	}
	return amount
}
