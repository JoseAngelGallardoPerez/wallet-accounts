package balance_test

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	. "github.com/Confialink/wallet-accounts/internal/modules/balance"
	mock_balance "github.com/Confialink/wallet-accounts/internal/modules/balance/mock"
	"time"

	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/shopspring/decimal"
)

var _ = Describe("Balance", func() {

	Context("Aggregation", func() {
		It("should ensure that default reducer produces correct result", func() {
			result := AggregationResult{
				{dec(-100), "EUR"},
				{dec(-10), "USD"},
				{dec(55.17), "CHF"},
				{dec(143.14), "BYN"},
				{dec(-12), "EUR"},
			}

			rateSource := exchange.NewDirectRateSource()
			_ = rateSource.Set(exchange.NewRate("EUR", "BTC", dec(0.000085)))
			_ = rateSource.Set(exchange.NewRate("USD", "BTC", dec(0.000073)))
			_ = rateSource.Set(exchange.NewRate("CHF", "BTC", dec(0.000080)))
			_ = rateSource.Set(exchange.NewRate("BYN", "BTC", dec(0.000028)))
			expectedTotal := dec(-0.00182848)

			reducer := NewDefaultReducer(rateSource)

			resultItem, err := reducer.Reduce(result, "BTC")
			Expect(err).ToNot(HaveOccurred())
			Expect(resultItem.CurrencyCode()).To(Equal("BTC"))
			Expect(resultItem.Amount()).To(decEqual(expectedTotal))
		})

		It("should represent aggregationService", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			items := AggregationResult{
				{dec(100), "EUR"},
				{dec(100), "USD"},
			}
			mockAggregator := mock_balance.NewMockAggregator(ctrl)
			mockAggregator.
				EXPECT().
				Aggregate().
				Return(items, nil).
				AnyTimes()

			mockFactory := mock_balance.NewMockAggregationFactory(ctrl)
			mockFactory.
				EXPECT().
				GeneralTotalByUserId(gomock.Any()).
				Return(mockAggregator, nil)
			mockFactory.
				EXPECT().
				TotalDebitedByUserIdPerPeriod(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockAggregator, nil)

			rateSource := exchange.NewDirectRateSource()
			_ = rateSource.Set(exchange.NewRate("EUR", "USD", dec(1.17)))
			_ = rateSource.Set(exchange.NewRate("USD", "EUR", dec(0.86)))

			reducer := NewDefaultReducer(rateSource)

			srv := NewAggregationService(reducer, mockFactory)
			result, err := srv.GeneralTotalByUserId("user_id", "EUR")
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Amount()).To(decEqual(dec(186)))
			Expect(result.CurrencyCode()).To(Equal("EUR"))

			result, err = srv.TotalDebitedByUserPerPeriod("user_id", time.Now(), time.Now(), "USD")
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Amount()).To(decEqual(dec(217)))
			Expect(result.CurrencyCode()).To(Equal("USD"))
		})
	})
})

func dec(d interface{}) decimal.Decimal {
	var v decimal.Decimal
	switch d := d.(type) {
	case decimal.Decimal:
		v = d
	case *decimal.Decimal:
		v = *d
	case string:
		v = str2Dec(d)
	case int:
		v = decimal.NewFromInt(int64(d))
	case int64:
		v = decimal.NewFromInt(d)
	case float32:
		v = decimal.NewFromFloat32(d)
	case float64:
		v = decimal.NewFromFloat(d)
	default:
		panic("invalid argument type")
	}
	return v
}

func str2Dec(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}

func decEqual(d interface{}) types.GomegaMatcher {
	return &decMatcher{
		Expected: dec(d),
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
