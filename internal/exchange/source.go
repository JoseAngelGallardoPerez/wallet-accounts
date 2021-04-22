package exchange

// RateSource
type RateSource interface {
	FindRate(base, reference string) (Rate, error)
}

type RateReceiver interface {
	Set(rate Rate) error
}

type RateSourceAndReceiver interface {
	RateSource
	RateReceiver
}
