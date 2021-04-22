package transfer

import (
	"github.com/shopspring/decimal"
)

// SelectAction is used in order to determine perform debit or credit in dependence of the given amount
type SelectAction struct {
	selector        ActionSelector
	performedAction Action
}

// ActionSelector is used in order to get action to perform
type ActionSelector interface {
	Select() (Action, error)
}

// NewSelectAction is SelectAction constructor
func NewSelectAction(selector ActionSelector) Action {
	return &SelectAction{selector: selector}
}

// Currency returns performed action currency or empty currency
func (s *SelectAction) Currency() Currency {
	if s.performedAction != nil {
		return s.performedAction.Currency()
	}
	return Currency{}
}

// Amount returns amount of performed action or zero
func (s *SelectAction) Amount() decimal.Decimal {
	if s.performedAction != nil {
		return s.performedAction.Amount()
	}
	return decimal.NewFromInt(0)
}

// Perform selects action and performs it
func (s *SelectAction) Perform() error {
	if s.performedAction != nil {
		return s.performedAction.Perform()
	}
	action, err := s.selector.Select()
	s.performedAction = action
	if err != nil {
		return err
	}
	return action.Perform()
}

// IsPerformed indicated whether selected action is performed
func (s *SelectAction) IsPerformed() bool {
	return s.performedAction != nil && s.performedAction.IsPerformed()
}

// Purpose returns selected action purpose or empty string
func (s *SelectAction) Purpose() string {
	if s.performedAction != nil {
		return s.performedAction.Purpose()
	}
	return ""
}

// Message returns selected action message or empty string
func (s *SelectAction) Message() string {
	if s.performedAction != nil {
		return s.performedAction.Message()
	}
	return ""
}

// selectorByAmountSign is used in order to select action based on the given amount
type selectorByAmountSign struct {
	actionA Action
	actionB Action
	amount  Amountable
}

func SelectByAmountSign(actionA Action, actionB Action, amount Amountable) ActionSelector {
	return &selectorByAmountSign{actionA: actionA, actionB: actionB, amount: amount}
}

// Select returns first action if amount is greater than zero, otherwise it returns the second action
func (s *selectorByAmountSign) Select() (Action, error) {
	if s.amount.Amount().GreaterThan(decimal.NewFromInt(0)) {
		return s.actionA, nil
	}
	return s.actionB, nil
}

// Sign indicates whether action is credit or debit
func (s *SelectAction) Sign() int {
	if s.performedAction == nil {
		return 0
	}
	return s.performedAction.Sign()
}
