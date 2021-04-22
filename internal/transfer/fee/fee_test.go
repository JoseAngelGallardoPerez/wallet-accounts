package fee_test

import (
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	. "github.com/onsi/gomega"

	. "github.com/Confialink/wallet-accounts/internal/transfer/fee"
)

var (
	decimal10                 = decimal.New(1, 1)
	decimal100                = decimal.New(1, 2)
	euroWallet, usdWallet     *transfer.Wallet
	euroCurrency, usdCurrency = *transfer.NewCurrency("EUR", 2), *transfer.NewCurrency("USD", 2)
)

var _ = Describe("Fee", func() {
	BeforeEach(func() {
		euroWallet = transfer.NewWallet(transfer.NewSimpleBalance(decimal100), euroCurrency)
		usdWallet = transfer.NewWallet(transfer.NewSimpleBalance(decimal100), usdCurrency)
	})
	Context("Transfer Fee", func() {
		It("should raise error in case if debitable and amount use different currencies", func() {
			params := TransferFeeParams{}
			_, err := NewDebitTransferFeeAction(euroWallet, transfer.NewAmount(usdCurrency, decimal10), params)
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(transfer.ErrCurrenciesMismatch))
		})

		Specify("transfer fee must be in the same currency as debitable", func() {
			params := TransferFeeParams{}
			transferFee, _ := NewDebitTransferFeeAction(euroWallet, transfer.NewAmount(euroCurrency, decimal10), params)
			Expect(transferFee.Currency()).To(Equal(euroWallet.Currency()))
		})
		It("should charge fixed amount", func() {
			params := TransferFeeParams{
				Base: str2Dec("0.3"),
			}
			debit, _ := transfer.NewDebitAction(euroWallet, transfer.NewAmount(euroCurrency, decimal10))
			transferFee, _ := NewDebitTransferFeeAction(euroWallet, debit, params)

			err := transfer.NewPerformerGroup(debit, transferFee).Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("89.7")))
			Expect(transferFee.IsPerformed()).To(BeTrue())
			Expect(transferFee.Amount()).To(decEqual(str2Dec("0.3")))
		})
		It("should charge 10% only", func() {
			params := TransferFeeParams{
				Percent: decimal10,
			}
			debit, _ := transfer.NewDebitAction(euroWallet, transfer.NewAmount(euroCurrency, decimal10))
			transferFee, _ := NewDebitTransferFeeAction(euroWallet, debit, params)

			err := transfer.NewPerformerGroup(debit, transferFee).Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("89")))
			Expect(transferFee.IsPerformed()).To(BeTrue())
			Expect(transferFee.Amount()).To(decEqual(str2Dec("1")))
		})
		It("should charge 10% + base fee", func() {
			params := TransferFeeParams{
				Percent: decimal10,
				Base:    str2Dec("0.3"),
			}
			debit, _ := transfer.NewDebitAction(euroWallet, transfer.NewAmount(euroCurrency, decimal10))
			transferFee, _ := NewDebitTransferFeeAction(euroWallet, debit, params)

			err := transfer.NewPerformerGroup(debit, transferFee).Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("88.7")))
			Expect(transferFee.IsPerformed()).To(BeTrue())
			Expect(transferFee.Amount()).To(decEqual(str2Dec("1.3")))
		})
		It("should not charge more than 1 + base fee", func() {
			params := TransferFeeParams{
				Percent: str2Dec("50"),
				Base:    str2Dec("0.3"),
				Max:     str2Dec("1"),
			}
			debit, _ := transfer.NewDebitAction(euroWallet, transfer.NewAmount(euroCurrency, decimal10))
			transferFee, _ := NewDebitTransferFeeAction(euroWallet, debit, params)

			err := transfer.NewPerformerGroup(debit, transferFee).Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("88.7")))
			Expect(transferFee.IsPerformed()).To(BeTrue())
			Expect(transferFee.Amount()).To(decEqual(str2Dec("1.3")))
		})
		It("should charge more than 1 + base fee but not less than 3 + base fee", func() {
			params := TransferFeeParams{
				Percent: str2Dec("50"),
				Base:    str2Dec("0.3"),
				Max:     str2Dec("3"),
				Min:     str2Dec("1"),
			}
			debit, _ := transfer.NewDebitAction(euroWallet, transfer.NewAmount(euroCurrency, decimal10))
			transferFee, _ := NewDebitTransferFeeAction(euroWallet, debit, params)

			err := transfer.NewPerformerGroup(debit, transferFee).Perform()
			Expect(err).ToNot(HaveOccurred())
			Expect(euroWallet.Amount()).To(decEqual(str2Dec("86.7")))
			Expect(transferFee.IsPerformed()).To(BeTrue())
			Expect(transferFee.Amount()).To(decEqual(str2Dec("3.3")))
		})
		When("percent is not specified", func() {
			Specify("min and max do not apply", func() {
				params := TransferFeeParams{
					Base: str2Dec("0.3"),
					Max:  str2Dec("3"),
					Min:  str2Dec("1"),
				}
				debit, _ := transfer.NewDebitAction(euroWallet, transfer.NewAmount(euroCurrency, decimal10))
				transferFee, _ := NewDebitTransferFeeAction(euroWallet, debit, params)

				err := transfer.NewPerformerGroup(debit, transferFee).Perform()
				Expect(err).ToNot(HaveOccurred())
				Expect(euroWallet.Amount()).To(decEqual(str2Dec("89.7")))
				Expect(transferFee.IsPerformed()).To(BeTrue())
				Expect(transferFee.Amount()).To(decEqual(str2Dec("0.3")))
			})
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
