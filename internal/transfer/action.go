package transfer

type Performer interface {
	// Perform executes expected logic
	Perform() error
	// IsPerformed indicates whether action is performed
	IsPerformed() bool
}

// Action describes required operations that are used in order to create transfer action
type Action interface {
	CurrencyAmount
	Performer
	// Purpose describes action purpose
	Purpose() string
	// Message is optional piece of information that clarifies the action
	Message() string
	// Sign indicates whether action is debit (-1) or credit (1), (0) if action is not credit or debit
	Sign() int
}

// PurposeSetter is used in order to specify that instance accept custom purpose
type PurposeSetter interface {
	SetPurpose(purpose string)
}

// PurposeSetter is used in order to specify that instance accept custom message
type MessageSetter interface {
	SetMessage(message string)
}
