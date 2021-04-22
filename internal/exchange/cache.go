package exchange

import "github.com/pkg/errors"

// cacheSource is used in order to preserve rates found from the given source
type cacheSource struct {
	topSource RateSource
	cache     RateSourceAndReceiver
}

// NewCacheSource is cache source constructor, it wraps passed rate source with cache
func NewCacheSource(topSource RateSource) RateSource {
	return &cacheSource{
		topSource: topSource,
		cache:     NewDirectRateSource(),
	}
}

// FindRate tries to fetch source from cache first
func (c *cacheSource) FindRate(base, reference string) (Rate, error) {
	rate, err := c.cache.FindRate(base, reference)
	if err == nil {
		return rate, nil
	}
	nilRate := Rate{}
	rate, err = c.topSource.FindRate(base, reference)
	if err == nil {
		err = c.cache.Set(rate)
		if err != nil {
			return nilRate, errors.Wrap(err, "failed to set rate to cache")
		}
		return rate, nil
	}
	return nilRate, errors.Wrap(err, "failed to retrieve rate from source")
}
