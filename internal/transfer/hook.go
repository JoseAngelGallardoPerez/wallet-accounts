package transfer

import (
	"github.com/shopspring/decimal"
)

// HookAction is used when there is a need to do some logic when a given action is performed
type HookAction struct {
	topAction Action
	callback  func(Action) error
}

// NewHookAction is HookAction constructor
func NewHookAction(topAction Action, callback func(Action) error) *HookAction {
	return &HookAction{topAction: topAction, callback: callback}
}

// Currency returns underlying action currency
func (h *HookAction) Currency() Currency {
	return h.topAction.Currency()
}

// Amount returns underlying action amount
func (h *HookAction) Amount() decimal.Decimal {
	return h.topAction.Amount()
}

// Perform calls given callback
func (h *HookAction) Perform() error {
	return h.callback(h.topAction)
}

// Purpose returns underlying action status
func (h *HookAction) IsPerformed() bool {
	return h.topAction.IsPerformed()
}

// Purpose returns underlying action purpose
func (h *HookAction) Purpose() string {
	return h.topAction.Purpose()
}

// Sign returns underlying action message
func (h *HookAction) Message() string {
	return h.topAction.Message()
}

// Sign returns underlying action sign
func (h *HookAction) Sign() int {
	return h.topAction.Sign()
}

// SetMessage sets message to a given action in case if it implements MessageSetter
func (h *HookAction) SetMessage(message string) {
	if setter, ok := h.topAction.(MessageSetter); ok {
		setter.SetMessage(message)
	}
}

// SetPurpose sets purpose to a given action in case if it implements PurposeSetter
func (h *HookAction) SetPurpose(purpose string) {
	if setter, ok := h.topAction.(PurposeSetter); ok {
		setter.SetPurpose(purpose)
	}
}
