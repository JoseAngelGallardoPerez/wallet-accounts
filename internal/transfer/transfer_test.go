package transfer_test

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	"github.com/Confialink/wallet-accounts/pkg/decround"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	. "github.com/Confialink/wallet-accounts/internal/transfer"
)

var (
	decimal10  = decimal.New(1, 1)
	decimal100 = decimal.New(1, 2)
)

func wallet(currency string) *Wallet {
	return NewWallet(NewSimpleBalance(decimal100), *NewCurrency(currency, 2))
}

var _ = Describe("Transfer", func() {
	Context("CurrentBalance", func() {
		It("should perform basic operations", func() {
			type testData struct {
				input    string
				expected string
			}

			v, _ := decimal.NewFromString("100.0")
			b := NewSimpleBalance(v)

			Expect(b.Amount().String()).To(Equal("100"))

			addTests := []testData{
				{"0.000001", "100.000001"},
				{"0.1", "100.100001"},
				{"1", "101.100001"},
				{"10", "111.100001"},
				{"0.000000000000000001", "111.100001000000000001"},
				{"333.333", "444.433001000000000001"},
				{"-10", "434.433001000000000001"},
				{"-1", "433.433001000000000001"},
				{"-0.000000000000000001", "433.433001"},
			}
			subTests := make([]testData, 0, len(addTests)-1)

			// creates tests for "Sub" operation based on the addTests
			for i := len(addTests) - 1; i > 0; i-- {
				subTests = append(subTests, testData{addTests[i].input, addTests[i-1].expected})
			}

			for _, test := range addTests {
				func(test testData, b Balance) {
					v, _ := decimal.NewFromString(test.input)
					err := b.Add(v)

					Expect(err).To(Not(HaveOccurred()))
					Expect(b.Amount().String()).To(Equal(test.expected))
				}(test, b)
			}

			for _, test := range subTests {
				func(test testData, b Balance) {
					v, _ := decimal.NewFromString(test.input)
					err := b.Sub(v)

					Expect(err).ToNot(HaveOccurred())
					Expect(b.Amount().String()).To(Equal(test.expected))
				}(test, b)
			}

		})
	})

	Context("Linked CurrentBalance", func() {
		It("should reflect changes in both directions", func() {
			decVal := decimal.NewFromInt(1000)
			linkedBalance := NewLinkedBalance(&decVal)

			Expect(linkedBalance.Add(decimal10)).To(Succeed())
			Expect(linkedBalance.Amount()).To(decEqual(decVal))
			Expect(decVal).To(decEqual(str2Dec("1010")))
			decVal = decVal.Add(decimal100)
			Expect(linkedBalance.Amount()).To(decEqual(str2Dec("1110")))
			Expect(linkedBalance.Amount()).To(decEqual(decVal))

			Expect(linkedBalance.Sub(decimal10)).To(Succeed())
			Expect(linkedBalance.Amount()).To(decEqual(decVal))
			Expect(decVal).To(decEqual(str2Dec("1100")))
			decVal = decVal.Sub(decimal100)
			Expect(linkedBalance.Amount()).To(decEqual(str2Dec("1000")))
			Expect(linkedBalance.Amount()).To(decEqual(decVal))
		})
	})

	Context("Currency", func() {
		It("should represent currency", func() {
			c := NewCurrency("EUR", 2)

			Expect(c.Code()).To(Equal("EUR"))
			Expect(c.Fraction()).To(Equal(uint(2)))
			Expect(c.String()).To(Equal(c.Code()))
		})
		It("should provide currency by code", func() {
			source := NewDirectCurrencySource()

			type td struct {
				code     string
				fraction uint
			}
			testData := []td{
				{"EUR", 2},
				{"USD", 2},
				{"ETH", 18},
				{"BTC", 8},
			}
			for _, data := range testData {
				Expect(source.Add(*NewCurrency(data.code, data.fraction))).To(Succeed())
			}
			for _, data := range testData {
				func(data td) {
					curr, err := source.Get(data.code)
					Expect(err).ToNot(HaveOccurred())
					Expect(curr.Code()).To(Equal(data.code))
					Expect(curr.Fraction()).To(Equal(data.fraction))
				}(data)
			}
			_, err := source.Get("NA")
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrencyNotFound))
		})
	})

	var (
		euroWallet, usdWallet     *Wallet
		euroCurrency, usdCurrency = *NewCurrency("EUR", 2), *NewCurrency("USD", 2)
	)

	Context("Amount", func() {
		It("should represent amount", func() {
			amount := NewAmount(
				*NewCurrency("EUR", 2),
				decimal10,
			)

			cur := amount.Currency()
			Expect(cur.Code()).To(Equal("EUR"))
			Expect(amount.Amount()).To(Equal(decimal10))
		})
		It("should return absolute amount", func() {
			absAmount := NewAmountAbs(NewAmount(euroCurrency, str2Dec("-100")))
			Expect(absAmount.Currency()).To(Equal(euroCurrency))
			Expect(absAmount.Amount()).To(decEqual(decimal100))

			absAmount = NewAmountAbs(NewAmount(euroCurrency, decimal100))
			Expect(absAmount.Amount()).To(decEqual(decimal100))
		})
	})

	Context("Wallet", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
			usdWallet = NewWallet(NewSimpleBalance(decimal100), usdCurrency)
		})

		It("should be correct output", func() {
			Expect(euroWallet.String()).To(Equal("EUR 100.00"))
			Expect(euroWallet.CurrencyCode()).To(Equal("EUR"))
			Expect(euroWallet.Amount()).To(decEqual(decimal100))
		})

		It("should credit an amount in the same currency to the wallet", func() {
			err := usdWallet.Credit(NewAmount(usdCurrency, str2Dec("10.50")))
			Expect(err).ToNot(HaveOccurred())
			Expect(usdWallet.Amount()).To(decEqual(str2Dec("110.50")))
			Expect(usdWallet.String()).To(Equal("USD 110.50"))
		})

		It("should NOT credit an amount in the different currency to the wallet", func() {
			err := usdWallet.Credit(NewAmount(euroCurrency, str2Dec("10.50")))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))
			Expect(usdWallet.Amount()).To(decEqual(str2Dec("100.00")))
		})

		It("should debit an amount in the same currency from the wallet", func() {
			err := euroWallet.Debit(NewAmount(euroCurrency, str2Dec("10.55")))
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("89.45")))
			Expect(euroWallet.String()).To(Equal("EUR 89.45"))
		})

		It("should NOT debit an amount in the different currency from the wallet", func() {
			err := euroWallet.Debit(NewAmount(usdCurrency, str2Dec("10.55")))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("100")))
			Expect(euroWallet.String()).To(Equal("EUR 100.00"))
		})

		It("should be possible to debit smaller amount", func() {
			err := euroWallet.Debit(NewAmount(euroCurrency, str2Dec("0.001")))
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("99.999")))
			Expect(euroWallet.String()).To(Equal("EUR 100.00"))
		})

		It("should be possible to debit amount greater than balance amount", func() {
			err := euroWallet.Debit(NewAmount(euroCurrency, str2Dec("200")))
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("-100")))
			Expect(euroWallet.String()).To(Equal("EUR -100.00"))
		})
		It("should negotiate amount", func() {
			amount := NewAmount(usdCurrency, decimal100)
			negAmount := NewAmountNeg(amount)

			Expect(negAmount.Currency()).To(Equal(usdCurrency))
			Expect(negAmount.Amount()).To(decEqual(str2Dec("-100")))
		})
		When("amount is zero or negative", func() {
			It("should not be possible to pass it to debit or credit", func() {
				negTen := NewAmount(euroCurrency, str2Dec("-10"))
				zero := NewAmount(euroCurrency, str2Dec("0"))

				err := euroWallet.Credit(negTen)
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrInvalidAmount))
				err = euroWallet.Credit(zero)
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrInvalidAmount))

				err = euroWallet.Debit(negTen)
				Expect(errors.Cause(err)).To(Equal(ErrInvalidAmount))
				Expect(err).To(HaveOccurred())
				err = euroWallet.Debit(zero)
				Expect(errors.Cause(err)).To(Equal(ErrInvalidAmount))
				Expect(err).To(HaveOccurred())

				Expect(euroWallet.Amount()).To(decEqual(str2Dec("100")))
			})
		})

		When("you need to replicate credit/debit", func() {
			It("should be possible to join creditables", func() {
				w1 := NewWallet(NewSimpleBalance(decimal100), euroCurrency)
				w2 := NewWallet(NewSimpleBalance(decimal10), euroCurrency)

				joined, err := JoinCreditable(w1, w2)
				Expect(err).ToNot(HaveOccurred())
				Expect(joined.Credit(NewAmount(euroCurrency, decimal10))).To(Succeed())
				Expect(w1.Amount()).To(decEqual(str2Dec("110")))
				Expect(w2.Amount()).To(decEqual(str2Dec("20")))
			})
			It("should not allow to join creditables in different currencies", func() {
				_, err := JoinCreditable(euroWallet, usdWallet)
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))
			})
			It("should be possible to join debitables", func() {
				w1 := NewWallet(NewSimpleBalance(decimal100), euroCurrency)
				w2 := NewWallet(NewSimpleBalance(decimal10), euroCurrency)

				joined, err := JoinDebitable(w1, w2)
				Expect(err).ToNot(HaveOccurred())
				Expect(joined.Debit(NewAmount(euroCurrency, decimal10))).To(Succeed())
				Expect(w1.Amount()).To(decEqual(str2Dec("90")))
				Expect(w2.Amount()).To(decEqual(str2Dec("0")))
			})
			It("should not allow to join debitables in different currencies", func() {
				_, err := JoinDebitable(euroWallet, usdWallet)
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))
			})
			It("should always return smallest amount", func() {
				w1 := NewWallet(NewSimpleBalance(decimal100), euroCurrency)
				w2 := NewWallet(NewSimpleBalance(decimal10), euroCurrency)

				joined, _ := JoinDebitable(w1, w2)
				Expect(joined.Amount()).To(decEqual(decimal10))

				_ = w2.Credit(NewAmount(euroCurrency, decimal100))
				Expect(joined.Amount()).To(decEqual(decimal100))
			})
		})

		It("should simply ignore debit or credit operations", func() {
			w := NewNoOpWallet(usdCurrency)

			Expect(w.Currency()).To(Equal(usdCurrency))
			Expect(w.Amount()).To(decEqual(str2Dec("0")))

			err := w.Debit(NewAmount(euroCurrency, decimal10))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch), "NoOpWallet should still restrict operations to the only given currency")

			err = w.Credit(NewAmount(euroCurrency, decimal10))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch), "NoOpWallet should still restrict operations to the only given currency")

			Expect(w.Debit(NewAmount(usdCurrency, decimal10))).To(Succeed())
			Expect(w.Credit(NewAmount(usdCurrency, decimal100))).To(Succeed())
			Expect(w.Amount()).To(decEqual(str2Dec("0")))
		})
	})

	Context("Debit Action", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
		})

		It("should debit wallet", func() {
			debitAction, err := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(decimal100))
			Expect(debitAction.IsPerformed()).To(BeFalse())
			err = debitAction.Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("90")))
			Expect(debitAction.IsPerformed()).To(BeTrue())
			Expect(debitAction.Sign()).To(Equal(-1))

			err = debitAction.Perform()
			Expect(err).To(HaveOccurred()) // by default action can be performed only once
			Expect(errors.Cause(err)).To(Equal(ErrAlreadyPerformed))
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("90")))
		})

		It("should fail to create debit action with different currencies", func() {
			_, err := NewDebitAction(euroWallet, NewAmount(usdCurrency, decimal10))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))
		})

		Specify("debit action should return the same currency as wallet", func() {
			debitAction, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			cur := debitAction.Currency()
			walletCur := euroWallet.Currency()
			Expect(cur.Code()).To(Equal(walletCur.Code()))
		})

		When("debit action is not performed", func() {
			It("should return 0 amount", func() {
				debitAction, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
				Expect(debitAction.Amount()).To(decEqual(str2Dec("0")))
			})
		})

		When("debit action performed", func() {
			It("should return the same amount as debited", func() {
				debitAction, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
				_ = debitAction.Perform()
				Expect(debitAction.Amount()).To(decEqual(decimal10))
			})
		})

		When("purpose and message are NOT set", func() {
			It("should return default purpose and message", func() {
				debitAction, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
				Expect(debitAction.Purpose()).To(Equal("debit"))
				Expect(debitAction.Message()).To(Equal(""))
			})
		})

		When("purpose and message are set", func() {
			It("should return what was set", func() {
				debitAction, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
				purpose, msg := "debit 10 EUR from wallet X", "payment for the service Y"
				debitAction.SetPurpose(purpose)
				debitAction.SetMessage(msg)
				Expect(debitAction.Purpose()).To(Equal(purpose))
				Expect(debitAction.Message()).To(Equal(msg))
			})
		})

	})

	Context("CreditFromAlias Action", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
		})

		It("should credit wallet", func() {
			creditAction, err := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(decimal100))
			Expect(creditAction.IsPerformed()).To(BeFalse())
			err = creditAction.Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("110")))
			Expect(creditAction.IsPerformed()).To(BeTrue())
			Expect(creditAction.Sign()).To(Equal(1))

			err = creditAction.Perform()
			Expect(err).To(HaveOccurred()) // by default action can be performed only once
			Expect(errors.Cause(err)).To(Equal(ErrAlreadyPerformed))
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("110")))
		})

		It("should fail to create credit action with different currencies", func() {
			_, err := NewCreditAction(euroWallet, NewAmount(usdCurrency, decimal10))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))
		})

		Specify("credit action should return the same currency as wallet", func() {
			creditAction, _ := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
			cur := creditAction.Currency()
			walletCur := euroWallet.Currency()
			Expect(cur.Code()).To(Equal(walletCur.Code()))
		})

		When("credit action is not performed", func() {
			It("should return 0 amount", func() {
				creditAction, _ := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
				Expect(creditAction.Amount()).To(decEqual(str2Dec("0")))
			})
		})

		When("credit action performed", func() {
			It("should return the same amount as credited", func() {
				creditAction, _ := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
				_ = creditAction.Perform()
				Expect(creditAction.Amount()).To(decEqual(decimal10))
			})
		})

		When("purpose and message are NOT set", func() {
			It("should return default purpose and message", func() {
				creditAction, _ := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
				Expect(creditAction.Purpose()).To(Equal("credit"))
				Expect(creditAction.Message()).To(Equal(""))
			})
		})

		When("purpose and message are set", func() {
			It("should return what was set", func() {
				creditAction, _ := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
				purpose, msg := "credit 10 EUR from wallet X", "payout for the service Y"
				creditAction.SetPurpose(purpose)
				creditAction.SetMessage(msg)
				Expect(creditAction.Purpose()).To(Equal(purpose))
				Expect(creditAction.Message()).To(Equal(msg))
			})
		})

	})

	Context("Action selector", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
		})
		It("should select credit action", func() {
			credit, _ := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
			debit, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			credit.SetPurpose("custom")
			credit.SetMessage("custom")

			amount := NewSimpleBalance(str2Dec("10"))
			selectAction := NewSelectAction(SelectByAmountSign(credit, debit, amount))
			Expect(selectAction.Purpose()).To(Equal(""))
			Expect(selectAction.Message()).To(Equal(""))
			Expect(selectAction.Amount()).To(decEqual(str2Dec("0")))
			Expect(selectAction.Sign()).To(Equal(0))

			err := selectAction.Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(selectAction.IsPerformed()).To(BeTrue())
			Expect(selectAction.Sign()).To(Equal(1))
			Expect(credit.IsPerformed()).To(BeTrue())
			Expect(debit.IsPerformed()).To(BeFalse())
			Expect(selectAction.Purpose()).To(Equal(credit.Purpose()))
			Expect(selectAction.Message()).To(Equal(credit.Message()))
			Expect(selectAction.Amount()).To(decEqual(credit.Amount()))
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("110")))
		})
		It("should select debit action", func() {
			credit, _ := NewCreditAction(euroWallet, NewAmount(euroCurrency, decimal10))
			debit, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			debit.SetPurpose("custom")
			debit.SetMessage("custom")

			amount := NewSimpleBalance(str2Dec("-10"))
			selectAction := NewSelectAction(SelectByAmountSign(credit, debit, amount))
			Expect(selectAction.Purpose()).To(Equal(""))
			Expect(selectAction.Message()).To(Equal(""))
			Expect(selectAction.Amount()).To(decEqual(str2Dec("0")))
			Expect(selectAction.Sign()).To(Equal(0))

			err := selectAction.Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(selectAction.IsPerformed()).To(BeTrue())
			Expect(selectAction.Sign()).To(Equal(-1))
			Expect(credit.IsPerformed()).To(BeFalse())
			Expect(debit.IsPerformed()).To(BeTrue())
			Expect(selectAction.Purpose()).To(Equal(debit.Purpose()))
			Expect(selectAction.Message()).To(Equal(debit.Message()))
			Expect(selectAction.Amount()).To(decEqual(debit.Amount()))
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("90")))
		})
	})

	Context("Action hook", func() {
		It("should invoke given func", func() {
			someRunTimeInfo := ""

			creditAction, _ := NewCreditAction(wallet("EUR"), NewAmount(euroCurrency, decimal100))
			creditAction.SetPurpose("test")
			hookAction := NewHookAction(creditAction, func(action Action) error {
				err := action.Perform()
				curr := action.Currency()
				someRunTimeInfo = fmt.Sprintf(
					"Purpose: %s, Code: %s, Amount: %s",
					action.Purpose(),
					curr.Code(),
					action.Amount(),
				)
				return err
			})

			Expect(hookAction.Perform()).To(Succeed())
			Expect(someRunTimeInfo).To(Equal("Purpose: test, Code: EUR, Amount: 100"))

			hookAction.SetPurpose("custom")
			hookAction.SetMessage("custom")
			Expect(creditAction.Purpose()).To(Equal("custom"))
			Expect(creditAction.Message()).To(Equal("custom"))
			Expect(creditAction.Purpose()).To(Equal(hookAction.Purpose()))
			Expect(creditAction.Message()).To(Equal(hookAction.Message()))

		})
	})

	Context("Pre-Transfer", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
		})
		It("should credit the amount debited by another action", func() {
			euroWallet2 := NewWallet(NewSimpleBalance(decimal10), euroCurrency)
			debit, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			credit, _ := NewCreditAction(euroWallet2, debit)

			Expect(debit.Perform()).To(Succeed())
			Expect(credit.Perform()).To(Succeed())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("90")))
			Expect(euroWallet2.Amount()).To(decEqual(str2Dec("20")))
		})
		It("should not credit funds if debit is not performed", func() {
			euroWallet2 := NewWallet(NewSimpleBalance(decimal10), euroCurrency)
			debit, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			credit, _ := NewCreditAction(euroWallet2, debit)

			Expect(credit.Perform()).To(Succeed()) // succeed to perform the action but not credit funds
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("100")))
			Expect(euroWallet2.Amount()).To(decEqual(str2Dec("10")))
		})
	})

	Context("Parts", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
		})
		It("should correctly calculate multiplication", func() {
			type testData struct {
				input      string
				multiplier string
				expected   string
			}
			tests := []testData{
				{"1", "1", "1"},
				{"1", "2", "2"},
				{"1", "1.5", "1.5"},
				{"1", "1.00000001", "1.00000001"},
				{"1", ".00000000000001", ".00000000000001"},
				{"100", "0.3", "30"},
				{"100", "0.78", "78"},
			}

			for _, test := range tests {
				tests = append(
					tests,
					testData{test.input, "-" + test.multiplier, "-" + test.expected},
					testData{"-" + test.input, "-" + test.multiplier, test.expected},
					testData{"-" + test.input, test.multiplier, "-" + test.expected},
				)
			}
			for _, test := range tests {
				func(test testData) {
					amount := NewAmount(euroCurrency, str2Dec(test.input))
					mul := NewAmountMultiplier(amount, str2Dec(test.multiplier))
					Expect(mul.Amount()).To(decEqual(str2Dec(test.expected)))
				}(test)
			}
		})
		It("should debit fee from the same wallet", func() {
			destinationWallet := NewWallet(NewSimpleBalance(str2Dec("0")), euroCurrency)
			feeWallet := NewWallet(NewSimpleBalance(str2Dec("0")), euroCurrency)

			debit, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			// take 30% fee
			fee, _ := NewDebitAction(euroWallet, NewAmountMultiplier(debit, str2Dec("0.3")))
			creditFee, _ := NewCreditAction(feeWallet, fee)
			creditDestination, _ := NewCreditAction(destinationWallet, debit)

			for _, action := range []Action{debit, fee, creditFee, creditDestination} {
				err := action.Perform()
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("87")))
			Expect(feeWallet.Amount()).To(decEqual(str2Dec("3")))
			Expect(creditDestination.Amount()).To(decEqual(str2Dec("10")))
		})
		It("should consume from amount", func() {
			amountToSpend := NewAmountConsumable(NewAmount(euroCurrency, str2Dec("1000")))
			Expect(amountToSpend.Currency()).To(Equal(euroCurrency))
			Expect(amountToSpend.Amount()).To(decEqual(str2Dec("1000")))

			err := amountToSpend.Debit(NewAmount(euroCurrency, str2Dec("500")))
			Expect(err).ToNot(HaveOccurred())
			Expect(amountToSpend.Amount()).To(decEqual(str2Dec("500")))
			// not enough funds
			err = amountToSpend.Debit(NewAmount(euroCurrency, str2Dec("600")))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrNotEnoughFunds))
			// can't be debited with different currency
			err = amountToSpend.Debit(NewAmount(usdCurrency, str2Dec("300")))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))
			// can't be debited negative amount
			err = amountToSpend.Debit(NewAmount(euroCurrency, str2Dec("-1")))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrInvalidAmount))
			err = amountToSpend.Debit(NewAmount(euroCurrency, str2Dec("500")))
			Expect(err).ToNot(HaveOccurred())
			Expect(amountToSpend.Amount()).To(decEqual(str2Dec("0")))

		})
		It("should debit wallet and split the amount between 2 destinations (fee included in source debit)", func() {
			destinationWallet := NewWallet(NewSimpleBalance(str2Dec("0")), euroCurrency)
			feeWallet := NewWallet(NewSimpleBalance(str2Dec("0")), euroCurrency)

			debit, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal100))
			debitConsumable := NewAmountConsumable(debit)
			// take 30% fee
			fee, _ := NewDebitAction(debitConsumable, NewAmountMultiplier(debit, str2Dec("0.3")))
			creditFee, _ := NewCreditAction(feeWallet, fee)
			creditDestination, _ := NewCreditAction(destinationWallet, debitConsumable)

			for _, action := range []Action{debit, fee, creditFee, creditDestination} {
				err := action.Perform()
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("0")))
			Expect(feeWallet.Amount()).To(decEqual(str2Dec("30")))
			Expect(creditDestination.Amount()).To(decEqual(str2Dec("70")))

		})
	})

	directRate := exchange.NewDirectRateSource()
	_ = directRate.Set(exchange.NewRate("EUR", "USD", str2Dec("1.25")))
	rateSource := exchange.NewReverseRateSource(directRate)

	Context("Different currencies", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
			usdWallet = NewWallet(NewSimpleBalance(decimal100), usdCurrency)
		})
		When("exchange rate is specified", func() {
			It("should be possible to debit/credit amounts in different currencies", func() {
				debit10Eur, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, str2Dec("10")))
				exchangeEurToUsd := NewExchangeAction(debit10Eur, rateSource, usdCurrency)
				credit12Usd, _ := NewCreditAction(usdWallet, exchangeEurToUsd)
				for _, action := range []Action{debit10Eur, exchangeEurToUsd, credit12Usd} {
					err := action.Perform()
					Expect(err).ToNot(HaveOccurred())
				}
				Expect(euroWallet.Amount()).To(decEqual(str2Dec("90")))
				exchangeCur := exchangeEurToUsd.Currency()
				Expect(exchangeCur).To(Equal(usdCurrency))
				Expect(exchangeEurToUsd.Sign()).To(Equal(0))
				Expect(exchangeEurToUsd.Amount()).To(decEqual(str2Dec("12.5")))
				Expect(usdWallet.Amount()).To(decEqual(str2Dec("112.5")))
			})
			It("should be possible to specify desired amount in destination currency", func() {
				// desired amount is 25 USD
				amountInUsd := NewAmount(usdCurrency, str2Dec("25"))
				exchangeUsdToEur := NewExchangeAction(amountInUsd, rateSource, euroCurrency)

				debitFromEur, _ := NewDebitAction(euroWallet, exchangeUsdToEur)
				creditToUsd, _ := NewCreditAction(usdWallet, amountInUsd)
				for _, action := range []Action{exchangeUsdToEur, debitFromEur, creditToUsd} {
					err := action.Perform()
					Expect(err).ToNot(HaveOccurred())
				}
				Expect(exchangeUsdToEur.Amount()).To(decEqual(str2Dec("20")))
				Expect(usdWallet.Amount()).To(decEqual(str2Dec("125")))
				Expect(euroWallet.Amount()).To(decEqual(str2Dec("80")))
			})
		})
	})

	Context("Rounding", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
			usdWallet = NewWallet(NewSimpleBalance(decimal100), usdCurrency)
		})
		When("some currency amount does not fit into the currency fraction", func() {
			It("should be possible to split the amount so that one part fits the fraction", func() {
				amountInEur := NewAmount(euroCurrency, str2Dec("10.1012345"))
				amountInEur, remainderInEur := NewRoundAmount(amountInEur, decround.Truncate)

				Expect(amountInEur.Amount()).To(decEqual(str2Dec("10.10")))
				Expect(remainderInEur.Amount()).To(decEqual(str2Dec("0.0012345")))
			})
		})
		When("some currency amount fits the currency fraction", func() {
			Specify("reminder must be zero", func() {
				amountInEur := NewAmount(euroCurrency, str2Dec("10.10"))
				amountInEur, remainderInEur := NewRoundAmount(amountInEur, decround.Truncate)

				Expect(amountInEur.Amount()).To(decEqual(str2Dec("10.10")))
				Expect(remainderInEur.Amount()).To(decEqual(str2Dec("0")))
			})
		})
		When("amount is rounded", func() {
			It("should be possible to accumulate difference on gain/loss wallets", func() {
				gainLossEurWallet := NewWallet(NewSimpleBalance(str2Dec("0")), euroCurrency)
				amountInUsd := NewAmount(usdCurrency, str2Dec("7.56"))

				debitUsd, _ := NewDebitAction(usdWallet, amountInUsd)
				exchangeToEur := NewExchangeAction(debitUsd, rateSource, euroCurrency)

				exchangedEur, exchangedEurRem := NewRoundAmount(exchangeToEur, decround.Truncate)

				creditToEurWallet, _ := NewCreditAction(euroWallet, exchangedEur)
				creditToGainLossWallet, _ := NewCreditAction(gainLossEurWallet, exchangedEurRem)

				for _, action := range []Action{debitUsd, exchangeToEur, creditToEurWallet, creditToGainLossWallet} {
					err := action.Perform()
					Expect(err).ToNot(HaveOccurred())
				}

				Expect(debitUsd.Amount()).To(decEqual(str2Dec("7.56")))
				Expect(exchangeToEur.Amount()).To(decEqual(str2Dec("6.048")))
				Expect(exchangedEur.Amount()).To(decEqual(str2Dec("6.04")))
				Expect(exchangedEurRem.Amount()).To(decEqual(str2Dec("0.008")))
				Expect(gainLossEurWallet.Amount()).To(decEqual(str2Dec("0.008")))
			})
			Specify("remainder could be negative value", func() {
				gainLossEurWallet := NewWallet(NewSimpleBalance(str2Dec("0")), euroCurrency)
				amountInUsd := NewAmount(usdCurrency, str2Dec("7.56"))

				debitUsd, _ := NewDebitAction(usdWallet, amountInUsd)
				exchangeToEur := NewExchangeAction(debitUsd, rateSource, euroCurrency)

				exchangedEur, exchangedEurRem := NewRoundAmount(exchangeToEur, decround.HalfUp)

				creditToEurWallet, _ := NewCreditAction(euroWallet, exchangedEur)

				creditGainLoss, _ := NewCreditAction(gainLossEurWallet, exchangedEurRem)
				debitGainLoss, _ := NewDebitAction(gainLossEurWallet, NewAmountAbs(exchangedEurRem))
				// if exchangedEurRem.Amount() is greater than 0 then it takes first action,
				// otherwise it should take the second one
				debitOrCreditGainLossWallet := NewSelectAction(
					SelectByAmountSign(
						creditGainLoss,
						debitGainLoss,
						exchangedEurRem,
					),
				)

				for _, action := range []Action{debitUsd, exchangeToEur, creditToEurWallet, debitOrCreditGainLossWallet} {
					err := action.Perform()
					Expect(err).ToNot(HaveOccurred())
				}

				Expect(debitUsd.Amount()).To(decEqual(str2Dec("7.56")))
				Expect(exchangeToEur.Amount()).To(decEqual(str2Dec("6.048")))
				Expect(exchangedEur.Amount()).To(decEqual(str2Dec("6.05")))
				Expect(exchangedEurRem.Amount()).To(decEqual(str2Dec("-0.002")))
				Expect(gainLossEurWallet.Amount()).To(decEqual(str2Dec("-0.002")))
			})
		})
	})

	Context("Grouping", func() {
		BeforeEach(func() {
			euroWallet = NewWallet(NewSimpleBalance(decimal100), euroCurrency)
			usdWallet = NewWallet(NewSimpleBalance(decimal100), usdCurrency)
		})

		It("should be possible to group number of action", func() {
			debitEur, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			exchangeEurToUsd := NewExchangeAction(debitEur, rateSource, usdCurrency)
			creditUsd, _ := NewCreditAction(usdWallet, exchangeEurToUsd)

			group := NewPerformerGroup(debitEur, exchangeEurToUsd, creditUsd)
			err := group.Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(group.IsPerformed()).To(BeTrue())

			for _, action := range []Action{debitEur, exchangeEurToUsd, creditUsd} {
				Expect(action.IsPerformed()).To(BeTrue())
			}
		})
		It("should stop executing performers if there is an error", func() {
			emptyRates := exchange.NewDirectRateSource()
			debitEur, _ := NewDebitAction(euroWallet, NewAmount(euroCurrency, decimal10))
			exchangeEurToUsd := NewExchangeAction(debitEur, emptyRates, usdCurrency)
			creditUsd, _ := NewCreditAction(usdWallet, exchangeEurToUsd)

			group := NewPerformerGroup(debitEur, exchangeEurToUsd, creditUsd)
			err := group.Perform()
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(exchange.ErrRateNotFound))
			Expect(group.IsPerformed()).To(BeTrue())

			Expect(debitEur.IsPerformed()).To(BeTrue())
			Expect(exchangeEurToUsd.IsPerformed()).To(BeTrue()) // performed with error
			Expect(creditUsd.IsPerformed()).To(BeFalse())       // expected that after the error next action is not taken
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
