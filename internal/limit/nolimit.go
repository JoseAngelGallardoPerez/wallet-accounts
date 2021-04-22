package limit

// noLimit implements not limited behavior
type noLimit struct{}

func (n *noLimit) Available() Value {
	return n
}

func (n noLimit) WithinLimit(CurrencyAmount) error {
	return nil
}

func (n noLimit) NoLimit() bool {
	return true
}

func (n noLimit) CurrencyAmount() CurrencyAmount {
	return nil
}

// NoLimit creates not limited value
func NoLimit() Value {
	return &noLimit{}
}
