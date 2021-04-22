package limit

// Factory defines minimum requirements for creating limits
type Factory interface {
	// CreateLimit creates Limit by the given model
	CreateLimit(model Model) Limit
}

// NewFactory creates default factory
func NewFactory() Factory {
	return factory{}
}

type factory struct{}

func (f factory) CreateLimit(model Model) Limit {
	return New(model.Value)
}

// New limit creates limit based on the given value
func New(value Value) Limit {
	if lim, ok := value.(Limit); ok {
		return lim
	}
	if value.NoLimit() {
		return &noLimit{}
	}
	available := value.CurrencyAmount()

	return &max{
		amount:       available.Amount(),
		currencyCode: available.CurrencyCode(),
	}
}
