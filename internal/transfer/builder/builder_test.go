package builder_test

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	. "github.com/Confialink/wallet-accounts/internal/transfer/builder"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/shopspring/decimal"
)

var (
	decimal10                 = decimal.New(1, 1)
	decimal100                = decimal.New(1, 2)
	euroCurrency, usdCurrency = *transfer.NewCurrency("EUR", 2), *transfer.NewCurrency("USD", 2)
)

func wallet(currency string, balance ...decimal.Decimal) *transfer.Wallet {
	b := decimal100
	if len(balance) > 0 {
		b = balance[0]
	}
	return transfer.NewWallet(transfer.NewSimpleBalance(b), *transfer.NewCurrency(currency, 2))
}

func feePercent(percent int64) DebitActionProvider {
	return func(debitable transfer.Debitable, amount transfer.CurrencyAmount) (transfer.Action, error) {
		feeAmount := transfer.NewAmountMultiplier(amount, decimal.NewFromInt(percent).Div(decimal100))
		return transfer.NewDebitAction(debitable, feeAmount)
	}
}

var _ = Describe("Transfer", func() {
	Context("Expression", func() {
		It("should be possible to start the transfer from different types", func() {
			input := []interface{}{
				"10",
				10,
				int64(10),
				10.00,
				decimal.NewFromInt(10),
				transfer.NewAmount(euroCurrency, decimal.NewFromInt(10)),
			}
			for _, v := range input {
				func(v interface{}) {
					debitableWallet := wallet("EUR")
					Expect(Debit(v).From(debitableWallet).Execute()).To(Succeed())
					Expect(debitableWallet.Amount()).To(decEqual(str2Dec("90")))

					creditableWallet := wallet("EUR")
					Expect(Credit(v).To(creditableWallet).Execute()).To(Succeed())
					Expect(creditableWallet.Amount()).To(decEqual(str2Dec("110")))
				}(v)
			}
		})
		It("should panic if input value is in different currency", func() {
			euroWallet := wallet("EUR")
			invoke := func() {
				Debit(transfer.NewAmount(usdCurrency, decimal.NewFromInt(10))).From(euroWallet)
			}
			Expect(invoke).Should(Panic())

			invoke = func() {
				Credit(transfer.NewAmount(usdCurrency, decimal.NewFromInt(10))).To(euroWallet)
			}
			Expect(invoke).Should(Panic())
		})
		It("should be possible to exchange currencies using alias", func() {
			sourceWallet := wallet("EUR")
			destinationWallet := wallet("USD")
			rates := exchange.NewDirectRateSource()
			_ = rates.Set(exchange.NewRate("EUR", "USD", str2Dec("1.1")))

			chain := New().
				Debit(10).
				From(sourceWallet).
				As("debit").
				ExchangeFromAlias("debit").
				Using(rates).
				ToCurrency(destinationWallet.Currency()).
				As("debitInUsd").
				CreditFromAlias("debitInUsd").
				To(destinationWallet)

			Expect(chain.Execute()).To(Succeed())
			Expect(sourceWallet.Amount()).To(decEqual(str2Dec("90")))
			Expect(destinationWallet.Amount()).To(decEqual(str2Dec("111")))
			Expect(len(chain.Actions())).To(Equal(3))
		})
		It("should be possible to exchange currencies", func() {
			sourceWallet := wallet("EUR")
			destinationWallet := wallet("USD")
			rates := exchange.NewDirectRateSource()
			_ = rates.Set(exchange.NewRate("EUR", "USD", str2Dec("1.1")))
			debitAmount := transfer.NewAmount(euroCurrency, decimal10)

			chain := New().
				Debit(debitAmount).
				From(sourceWallet).
				Exchange(debitAmount).
				Using(rates).
				ToCurrency(destinationWallet.Currency()).
				As("debitInUsd").
				CreditFromAlias("debitInUsd").
				To(destinationWallet)

			Expect(chain.Execute()).To(Succeed())
			Expect(sourceWallet.Amount()).To(decEqual(str2Dec("90")))
			Expect(destinationWallet.Amount()).To(decEqual(str2Dec("111")))
			Expect(len(chain.Actions())).To(Equal(3))
		})
		It("should be possible to start from Exchange", func() {
			destinationWallet := wallet("USD", str2Dec("0"))
			rates := exchange.NewDirectRateSource()
			_ = rates.Set(exchange.NewRate("EUR", "USD", str2Dec("1.1")))

			chain := Exchange(transfer.NewAmount(euroCurrency, decimal10)).
				Using(rates).
				ToCurrency(destinationWallet).
				As("USD").
				CreditFromAlias("USD").
				To(destinationWallet)

			Expect(chain.Execute()).To(Succeed())
			Expect(destinationWallet.Amount()).To(decEqual(str2Dec("11")))
		})
		When("specifying destination currency", func() {
			It("should be possible to use different inputs", func() {
				input := []interface{}{
					"anotherDebit",
					usdCurrency,
					&usdCurrency,
					transfer.NewAmount(usdCurrency, decimal10),
				}
				rates := exchange.NewDirectRateSource()
				_ = rates.Set(exchange.NewRate("EUR", "USD", str2Dec("1.1")))
				usdWallet := wallet("USD")

				for _, v := range input {
					func(v interface{}) {
						sourceWallet := wallet("EUR")
						destinationWallet := wallet("USD")

						chain := Debit(10).
							From(sourceWallet).
							As("debit").
							Debit(1).
							From(usdWallet).
							As("anotherDebit").
							ExchangeFromAlias("debit").
							Using(rates).
							ToCurrency(v).
							As("debitInUsd").
							CreditFromAlias("debitInUsd").
							To(destinationWallet)

						Expect(chain.Execute()).To(Succeed())
						Expect(sourceWallet.Amount()).To(decEqual(str2Dec("90")))
						Expect(destinationWallet.Amount()).To(decEqual(str2Dec("111")))
					}(v)
				}
			})
		})
		It("should build simple transfer with debit/credit actions starting from Debit", func() {
			eurRevenue := wallet("EUR")
			eurDestination := wallet("EUR")

			chain := Debit(100).
				From(wallet("EUR")).
				As("debit1").
				WithPurpose("custom purpose").
				WithMessage("custom message").
				Debit(feePercent(10)).
				FromAlias("debit1").
				WithMessage("fee test").
				WithPurpose("fee test").
				As("fee10%").
				CreditFromAlias("debit1").
				To(eurDestination).
				CreditFromAlias("fee10%").
				To(eurRevenue)

			Expect(len(chain.Actions())).To(Equal(4))

			Expect(chain.Execute()).To(Succeed())

			Expect(eurRevenue.Amount()).To(decEqual(str2Dec("110")))
			Expect(eurDestination.Amount()).To(decEqual(str2Dec("200")))

		})

		It("should just credit", func() {
			creditableWallet := wallet("USD")

			chain := Credit("500.99").
				To(creditableWallet)

			Expect(chain.Execute()).To(Succeed())
			Expect(creditableWallet.Amount()).To(decEqual(str2Dec("600.99")))
			Expect(len(chain.Actions())).To(Equal(1))
		})

		It("should hook actions", func() {
			type creditInfo struct {
				Amount              decimal.Decimal
				InternalDescription string
				ExternalDescription string
				CreditAction        transfer.Action
				DebitAction         transfer.Action
			}
			info := &creditInfo{ExternalDescription: "external"}

			chain := Debit(500).
				From(wallet("EUR")).
				As("d1").
				WithCallback(func(action transfer.Action) error {
					info.DebitAction = action
					return action.Perform()
				}).
				CreditFromAlias("d1").
				To(wallet("EUR")).
				WithCallback(func(action transfer.Action) error {
					err := action.Perform()
					info.InternalDescription = "internal"
					info.Amount = action.Amount()
					info.CreditAction = action
					return err
				})

			Expect(info.InternalDescription).To(Equal(""))
			Expect(info.Amount).To(decEqual(str2Dec("0")))
			Expect(chain.Execute()).To(Succeed())
			Expect(info.InternalDescription).To(Equal("internal"))
			Expect(info.Amount).To(decEqual(str2Dec("500")))
			Expect(info.CreditAction.Sign()).To(Equal(1))
			Expect(info.DebitAction.Sign()).To(Equal(-1))

		})

		It("should sum values by name", func() {
			w := wallet("USD")

			chain := Debit(10).
				From(w).
				IncludeToGroup("total").
				Debit(20).
				From(w).
				IncludeToGroup("total").
				Credit(5).
				To(w).
				IncludeToGroup("total")

			Expect(chain.GetGroup("total").Sum()).To(decEqual(str2Dec("0")))
			Expect(chain.GetGroup("total")).To(HaveLen(3))
			Expect(chain.Execute()).To(Succeed())
			// -10 -20 +5
			Expect(chain.GetGroup("total").Sum()).To(decEqual(str2Dec("-25")))
		})

		It("should retrieve amount by alias name", func() {
			chain1 := Debit(100).
				From(wallet("EUR")).
				As("alias")

			chain2 := Credit(100).
				To(wallet("EUR")).
				As("alias")

			rates := exchange.NewDirectRateSource()
			_ = rates.Set(exchange.NewRate("EUR", "USD", str2Dec("1.1")))
			chain3 := Exchange(transfer.NewAmount(euroCurrency, decimal100)).
				Using(rates).
				ToCurrency(usdCurrency).
				As("alias")

			type testData struct {
				t        *Transfer
				expected decimal.Decimal
			}

			for _, data := range []testData{
				{chain1, decimal100},
				{chain2, decimal100},
				{chain3, str2Dec("110")},
			} {
				func(data testData) {
					alias := data.t.AmountAlias("alias")
					Expect(alias).ToNot(BeNil())
					Expect(alias.Amount()).To(decEqual(str2Dec("0")))
					Expect(data.t.Execute()).To(Succeed())
					Expect(alias.Amount()).To(decEqual(data.expected))
					Expect(data.t.AmountAlias("unknown")).To(BeNil())
				}(data)
			}

		})

		It("should panic if unexpected value is given", func() {
			rates := exchange.NewDirectRateSource()
			_ = rates.Set(exchange.NewRate("EUR", "USD", str2Dec("1.1")))
			amount := transfer.NewAmount(euroCurrency, decimal100)
			unexpectedValues := []interface{}{
				nil,
				"not a number",
				" 123",
				"a",
			}
			funcs := make([]func(), 0, len(unexpectedValues)*3)
			for _, v := range unexpectedValues {
				funcs = append(funcs, func() {
					Debit(v).From(wallet("EUR"))
				})
				funcs = append(funcs, func() {
					Credit(v).To(wallet("EUR"))
				})
				funcs = append(funcs, func() {
					Exchange(amount).Using(rates).ToCurrency(v)
				})
			}

			for _, f := range funcs {
				func(func()) {
					Expect(f).Should(Panic())
				}(f)
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
