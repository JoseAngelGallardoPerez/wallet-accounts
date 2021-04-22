package currency_test

import (
	"github.com/Confialink/wallet-accounts/internal/modules/currency"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/model"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"fmt"
	"github.com/onsi/gomega/types"
	"github.com/shopspring/decimal"

	mock_service "github.com/Confialink/wallet-accounts/internal/modules/currency/service/mock"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Currency", func() {
	Context("Provider", func() {
		It("caches results from the given service", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			srv := mock_service.NewMockCurrenciesServiceInterface(ctrl)
			srv.
				EXPECT().
				GetByCode("EUR").
				Return(&model.Currency{
					Id:            1,
					Code:          "EUR",
					Active:        true,
					DecimalPlaces: 2,
				}, nil)

			provider := currency.NewProvider(srv)

			cur, err := provider.Get("EUR")
			Expect(err).ToNot(HaveOccurred())
			Expect(cur.Code()).To(Equal("EUR"))
			Expect(cur.Fraction()).To(BeEquivalentTo(2))

			_, err = provider.Get("EUR")
			Expect(err).ToNot(HaveOccurred())

		})

		It("returns error in case if the srv returns error", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			srv := mock_service.NewMockCurrenciesServiceInterface(ctrl)
			srv.
				EXPECT().
				GetByCode("EUR").
				Return(nil, errors.New("some error"))

			provider := currency.NewProvider(srv)

			cur, err := provider.Get("EUR")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("some error"))
			Expect(cur).To(BeIdenticalTo(transfer.Currency{}))
		})
	})
	Context("Rate Source", func() {
		It("should find rate using given currencies service", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			srv := mock_service.NewMockCurrenciesServiceInterface(ctrl)
			srv.
				EXPECT().
				GetCurrenciesRateByCodes("EUR", "USD").
				Return(&service.Rate{
					Rate:           decimal.NewFromFloat(1.11),
					ExchangeMargin: decimal.Zero,
				}, nil)
			srv.
				EXPECT().
				GetCurrenciesRateByCodes("UNK", "UNK").
				Return(&service.Rate{}, errors.New("some error"))

			source := currency.NewRateSource(srv)

			rate, err := source.FindRate("EUR", "USD")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rate.BaseCurrencyCode()).To(Equal("EUR"))
			Expect(rate.ReferenceCurrencyCode()).To(Equal("USD"))
			Expect(rate.Rate()).To(decEqual(str2Dec("1.11")))

			_, err = source.FindRate("UNK", "UNK")
			Expect(err).To(HaveOccurred())
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
