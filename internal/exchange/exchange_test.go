package exchange_test

import (
	. "github.com/Confialink/wallet-accounts/internal/exchange"
	mock_exchange "github.com/Confialink/wallet-accounts/internal/exchange/mock"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	decimal2 = decimal.NewFromInt(2)
	decimal4 = decimal.NewFromInt(4)
)

var _ = Describe("Rate", func() {
	Context("Checking rate", func() {
		r := One("USD").Is(decimal2, "EUR")
		It("should return correct codes", func() {
			Expect(r.BaseCurrencyCode()).To(Equal("USD"))
			Expect(r.ReferenceCurrencyCode()).To(Equal("EUR"))
		})
		It("should return correct rate", func() {
			Expect(r.Rate()).To(decEqual(str2Dec("2")))
		})
	})

	Context("Checking rate source", func() {
		It("should successfully set direct rates", func() {
			directSource := NewDirectRateSource()
			Expect(directSource.Set(NewRate("EUR", "USD", decimal2))).Should(Succeed())
			Expect(directSource.Set(NewRate("EUR", "CAD", decimal4))).Should(Succeed())
		})

		It("should only return directRateSource rates", func() {
			directSource := NewDirectRateSource()
			_ = directSource.Set(NewRate("EUR", "USD", decimal2))
			_ = directSource.Set(NewRate("USD", "JPY", decimal4))
			rate, err := directSource.FindRate("EUR", "USD")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("EUR"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("USD"))
			Expect(rate.Rate()).To(decEqual(str2Dec("2")))

			rate, err = directSource.FindRate("USD", "JPY")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("USD"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("JPY"))
			Expect(rate.Rate()).To(decEqual(str2Dec("4")))

			_, err = directSource.FindRate("EUR", "JPY")
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrRateNotFound))

			_, err = directSource.FindRate("USD", "EUR")
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrRateNotFound))
		})

		When("the same currency is requested", func() {
			It("should return rate(1) even if it was not set", func() {
				directSource := NewDirectRateSource()
				rate, err := directSource.FindRate("SAME", "SAME")
				Expect(err).ToNot(HaveOccurred())
				Expect(rate.Rate()).To(decEqual(str2Dec("1")))
			})
		})

		It("reverse source should preserve original source rates", func() {
			directSource := NewDirectRateSource()
			_ = directSource.Set(NewRate("EUR", "USD", decimal2))
			reverseSource := NewReverseRateSource(directSource)

			rate, err := reverseSource.FindRate("EUR", "USD")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("EUR"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("USD"))
			Expect(rate.Rate()).To(decEqual(str2Dec("2")))
		})

		It("reverse source should additionally allow to find reverse rates", func() {
			directSource := NewDirectRateSource()
			_ = directSource.Set(NewRate("EUR", "USD", decimal2))
			reverseSource := NewReverseRateSource(directSource)

			rate, err := reverseSource.FindRate("USD", "EUR")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("USD"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("EUR"))
			Expect(rate.Rate()).To(decEqual(str2Dec(".5")))
		})

		It("pivot source should preserve original source rates", func() {
			directSource := NewDirectRateSource()
			_ = directSource.Set(NewRate("EUR", "USD", decimal2))
			pivotSource := NewPivotRateSource("EUR", directSource)

			rate, err := pivotSource.FindRate("EUR", "USD")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("EUR"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("USD"))
			Expect(rate.Rate()).To(decEqual(str2Dec("2")))
		})

		It("should calculate rate based on pivot rate", func() {
			directSource := NewDirectRateSource()
			_ = directSource.Set(NewRate("EUR", "USD", decimal2))
			_ = directSource.Set(NewRate("EUR", "JPY", decimal4))

			pivotSource := NewPivotRateSource("EUR", directSource)
			rate, err := pivotSource.FindRate("USD", "JPY")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("USD"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("JPY"))
			Expect(rate.Rate()).To(decEqual(str2Dec("2")))

			rate, err = pivotSource.FindRate("JPY", "USD")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("JPY"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("USD"))
			Expect(rate.Rate()).To(decEqual(str2Dec(".5")))
		})

		It("should find pivot rate using reverse source", func() {
			directSource := NewDirectRateSource()
			_ = directSource.Set(NewRate("EUR", "USD", decimal2))
			_ = directSource.Set(NewRate("JPY", "EUR", decimal4))

			reverseSource := NewReverseRateSource(directSource)
			pivotSource := NewPivotRateSource("EUR", reverseSource)

			rate, err := pivotSource.FindRate("JPY", "USD")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("JPY"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("USD"))
			Expect(rate.Rate()).To(decEqual(str2Dec("8")))

			rate, err = pivotSource.FindRate("USD", "EUR")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("USD"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("EUR"))
			Expect(rate.Rate()).To(decEqual(str2Dec(".5")))
		})

		It("should preserve rate source results", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockSource := mock_exchange.NewMockRateSource(ctrl)
			mockSource.
				EXPECT().
				FindRate("EUR", "USD").
				Return(NewRate("EUR", "USD", str2Dec("1.11")), nil)

			mockSource.
				EXPECT().
				FindRate("EUR", "RUB").
				Return(NewRate("EUR", "RUB", str2Dec("87")), nil).
				After(
					mockSource.
						EXPECT().
						FindRate("EUR", "RUB").
						Return(Rate{}, ErrRateNotFound),
				)

			cached := NewCacheSource(mockSource)

			for i := 0; i < 2; i++ {
				func() {
					rate, err := cached.FindRate("EUR", "USD")
					Expect(err).ShouldNot(HaveOccurred())
					Expect(rate.BaseCurrencyCode()).To(Equal("EUR"))
					Expect(rate.ReferenceCurrencyCode()).To(Equal("USD"))
					Expect(rate.Rate()).To(decEqual(str2Dec("1.11")))
				}()
			}

			rate, err := cached.FindRate("EUR", "RUB")
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrRateNotFound))
			Expect(rate).To(Equal(Rate{}))

			for i := 0; i < 2; i++ {
				func() {
					rate, err := cached.FindRate("EUR", "RUB")
					Expect(err).ShouldNot(HaveOccurred())
					Expect(rate.BaseCurrencyCode()).To(Equal("EUR"))
					Expect(rate.ReferenceCurrencyCode()).To(Equal("RUB"))
					Expect(rate.Rate()).To(decEqual(str2Dec("87")))
				}()
			}

		})
	})
})

func str2Dec(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}

func decEqual(d decimal.Decimal) types.GomegaMatcher {
	return &decMatcher{
		Expected: d,
	}
}

type decMatcher struct {
	Expected decimal.Decimal
}

func (d decMatcher) Match(actual interface{}) (success bool, err error) {
	a := actual.(decimal.Decimal)
	return d.Expected.Equal(a), nil
}

func (d decMatcher) FailureMessage(actual interface{}) (message string) {
	a := actual.(decimal.Decimal)
	return fmt.Sprintf("Expected <Decimal> %s to be <Decimal> %s", a.String(), d.Expected.String())
}

func (d decMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	a := actual.(decimal.Decimal)
	return fmt.Sprintf("Expected the value NOT to be <Decimal> %s", a.String())
}
