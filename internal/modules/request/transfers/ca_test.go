package transfers_test

import (
	accountTypeModel "github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	. "github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/shopspring/decimal"
	"time"
)

var (
	decimal10                 = decimal.New(1, 1)
	decimal100                = decimal.New(1, 2)
	euroCurrency, usdCurrency = *transfer.NewCurrency("EUR", 2), *transfer.NewCurrency("USD", 2)
	currencyBox               = transfer.NewDirectCurrencySource()
)

var _ = Describe("Transfers", func() {
	var (
		mock sqlmock.Sqlmock
		gdb  *gorm.DB
	)
	_ = currencyBox.Add(euroCurrency)
	_ = currencyBox.Add(usdCurrency)
	Context("CA(Credit Account) Transfer", func() {
		BeforeEach(func() {
			var db *sql.DB
			var err error

			db, mock, err = sqlmock.New() // mock sql.DB
			Expect(err).ShouldNot(HaveOccurred())

			gdb, err = gorm.Open("mysql", db) // open gorm db
			Expect(err).ShouldNot(HaveOccurred())
		})
		AfterEach(func() {
			err := mock.ExpectationsWereMet() // make sure all expectations were met
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should check transfer evaluation without iwt fee and revenue", func() {
			acc := account("EUR", "1000")
			rev := revenueAccount("EUR", "0")
			input := NewCreditAccountInput(acc, false, false, rev, nil)
			unit := NewCreditAccount(gdb, input, currencyBox)

			Expect(unit.Transactions()).To(HaveLen(0))

			details, err := unit.Evaluate(request("100", "EUR"))
			Expect(unit.Transactions()).To(HaveLen(1))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).Should(HaveLen(1))
			Expect(details).Should(HaveKey(constants.PurposeCreditAccount))

			detail := details[constants.PurposeCreditAccount]
			Expect(detail.Amount).To(decEqual(decimal100))
			Expect(detail.CurrencyCode).To(Equal("EUR"))
			Expect(*detail.AccountId).To(BeEquivalentTo(999))
			Expect(detail.RevenueAccountId).To(BeNil())
			Expect(detail.Transaction).NotTo(BeNil())

			transaction := detail.Transaction
			Expect(transaction.Id).To(BeNil(), "transaction should not be saved")
			Expect(*transaction.Amount).To(decEqual(str2Dec("100")))
			Expect(*transaction.AccountId).To(BeEquivalentTo(999))
			Expect(transaction.RevenueAccountId).To(BeNil())
			Expect(transaction.CardId).To(BeNil())
			Expect(transaction.ShowAmount).To(BeNil())
			Expect(transaction.ShowAvailableBalanceSnapshot).To(BeNil())
			Expect(*transaction.AvailableBalanceSnapshot).To(decEqual(str2Dec("1100")))

			Expect(acc.Balance).To(decEqual(str2Dec("1100")))
			Expect(acc.AvailableAmount).To(decEqual(str2Dec("1100")))
			Expect(rev.Balance).To(decEqual(decimal.Zero))
			Expect(rev.AvailableAmount).To(decEqual(decimal.Zero))
		})

		It("should check transfer evaluation without iwt fee, using revenue account", func() {
			acc := account("EUR", "1000")
			rev := revenueAccount("EUR", "0")
			input := NewCreditAccountInput(acc, false, true, rev, nil)
			unit := NewCreditAccount(gdb, input, currencyBox)

			Expect(unit.Transactions()).To(HaveLen(0))

			details, err := unit.Evaluate(request("100", "EUR"))
			Expect(unit.Transactions()).To(HaveLen(2))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).Should(HaveLen(2))
			Expect(details).Should(HaveKey(constants.PurposeCreditAccount))
			Expect(details).Should(HaveKey(constants.PurposeDebitRevenue))

			debitDetail := details[constants.PurposeDebitRevenue]
			Expect(debitDetail.Amount).To(decEqual(str2Dec("-100")))
			Expect(debitDetail.CurrencyCode).To(Equal("EUR"))
			Expect(debitDetail.AccountId).To(BeNil())
			Expect(debitDetail.RevenueAccountId).NotTo(BeNil())
			Expect(*debitDetail.RevenueAccountId).To(BeEquivalentTo(456))
			Expect(debitDetail.Transaction).NotTo(BeNil())

			debitTx := debitDetail.Transaction
			Expect(debitTx.Id).To(BeNil(), "debitTx should not be saved")
			Expect(*debitTx.Amount).To(decEqual(str2Dec("-100")))
			Expect(debitTx.AccountId).To(BeNil())
			Expect(*debitTx.RevenueAccountId).To(BeEquivalentTo(456))
			Expect(debitTx.CardId).To(BeNil())
			Expect(debitTx.ShowAmount).To(BeNil())
			Expect(debitTx.ShowAvailableBalanceSnapshot).To(BeNil())
			Expect(*debitTx.AvailableBalanceSnapshot).To(decEqual(str2Dec("-100")))

			creditDetail := details[constants.PurposeCreditAccount]
			Expect(creditDetail.Amount).To(decEqual(decimal100))
			Expect(creditDetail.CurrencyCode).To(Equal("EUR"))
			Expect(*creditDetail.AccountId).To(BeEquivalentTo(999))
			Expect(creditDetail.RevenueAccountId).To(BeNil())
			Expect(creditDetail.Transaction).NotTo(BeNil())

			creditTx := creditDetail.Transaction
			Expect(creditTx.Id).To(BeNil(), "creditTx should not be saved")
			Expect(*creditTx.Amount).To(decEqual(str2Dec("100")))
			Expect(*creditTx.AccountId).To(BeEquivalentTo(999))
			Expect(creditTx.RevenueAccountId).To(BeNil())
			Expect(creditTx.CardId).To(BeNil())
			Expect(creditTx.ShowAmount).To(BeNil())
			Expect(creditTx.ShowAvailableBalanceSnapshot).To(BeNil())
			Expect(*creditTx.AvailableBalanceSnapshot).To(decEqual(str2Dec("1100")))

			Expect(acc.Balance).To(decEqual(str2Dec("1100")))
			Expect(acc.AvailableAmount).To(decEqual(str2Dec("1100")))
			Expect(rev.Balance).To(decEqual(str2Dec("-100")))
			Expect(rev.AvailableAmount).To(decEqual(str2Dec("-100")))
		})

		It("should check transfer evaluation WITH iwt fee", func() {
			acc := account("EUR", "1000")
			rev := revenueAccount("EUR", "0")

			feeParams := &fee.TransferFeeParams{
				Percent: str2Dec("10"),
			}

			input := NewCreditAccountInput(acc, true, true, rev, feeParams)
			unit := NewCreditAccount(gdb, input, currencyBox)

			Expect(unit.Transactions()).To(HaveLen(0))

			details, err := unit.Evaluate(request("100", "EUR"))
			Expect(unit.Transactions()).To(HaveLen(4))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).Should(HaveLen(4))
			Expect(details).Should(HaveKey(constants.PurposeCreditAccount))
			Expect(details).Should(HaveKey(constants.PurposeDebitRevenue))
			Expect(details).Should(HaveKey(constants.PurposeFeeIWT))
			Expect(details).Should(HaveKey(constants.PurposeRevenueIwt))

			debitDetail := details[constants.PurposeDebitRevenue]
			Expect(debitDetail.Amount).To(decEqual(str2Dec("-100")))
			Expect(debitDetail.CurrencyCode).To(Equal("EUR"))
			Expect(debitDetail.AccountId).To(BeNil())
			Expect(debitDetail.RevenueAccountId).NotTo(BeNil())
			Expect(*debitDetail.RevenueAccountId).To(BeEquivalentTo(456))
			Expect(debitDetail.Transaction).NotTo(BeNil())

			debitTx := debitDetail.Transaction
			Expect(debitTx.Id).To(BeNil(), "debitTx should not be saved")
			Expect(*debitTx.Amount).To(decEqual(str2Dec("-100")))
			Expect(debitTx.AccountId).To(BeNil())
			Expect(*debitTx.RevenueAccountId).To(BeEquivalentTo(456))
			Expect(debitTx.CardId).To(BeNil())
			Expect(debitTx.ShowAmount).To(BeNil())
			Expect(debitTx.ShowAvailableBalanceSnapshot).To(BeNil())
			Expect(*debitTx.AvailableBalanceSnapshot).To(decEqual(str2Dec("-100")))

			creditDetail := details[constants.PurposeCreditAccount]
			Expect(creditDetail.Amount).To(decEqual(decimal100))
			Expect(creditDetail.CurrencyCode).To(Equal("EUR"))
			Expect(*creditDetail.AccountId).To(BeEquivalentTo(999))
			Expect(creditDetail.RevenueAccountId).To(BeNil())
			Expect(creditDetail.Transaction).NotTo(BeNil())

			creditTx := creditDetail.Transaction
			Expect(creditTx.Id).To(BeNil(), "creditTx should not be saved")
			Expect(*creditTx.Amount).To(decEqual(str2Dec("100")))
			Expect(*creditTx.AccountId).To(BeEquivalentTo(999))
			Expect(creditTx.RevenueAccountId).To(BeNil())
			Expect(creditTx.CardId).To(BeNil())
			Expect(creditTx.ShowAmount).To(BeNil())
			Expect(creditTx.ShowAvailableBalanceSnapshot).To(BeNil())
			Expect(*creditTx.AvailableBalanceSnapshot).To(decEqual(str2Dec("1100")))

			debitFeeDetail := details[constants.PurposeFeeIWT]
			Expect(debitFeeDetail.Amount).To(decEqual(decimal10))
			Expect(debitFeeDetail.CurrencyCode).To(Equal("EUR"))
			Expect(*debitFeeDetail.AccountId).To(BeEquivalentTo(999))
			Expect(debitFeeDetail.RevenueAccountId).To(BeNil())
			Expect(debitFeeDetail.Transaction).NotTo(BeNil())

			debitFeeTx := debitFeeDetail.Transaction
			Expect(debitFeeTx.Id).To(BeNil(), "debitFeeTx should not be saved")
			Expect(*debitFeeTx.Amount).To(decEqual(str2Dec("-10")))
			Expect(*debitFeeTx.AccountId).To(BeEquivalentTo(999))
			Expect(debitFeeTx.RevenueAccountId).To(BeNil())
			Expect(debitFeeTx.CardId).To(BeNil())
			Expect(debitFeeTx.ShowAmount).To(BeNil())
			Expect(debitFeeTx.ShowAvailableBalanceSnapshot).To(BeNil())
			Expect(*debitFeeTx.AvailableBalanceSnapshot).To(decEqual(str2Dec("1090")))

			creditFeeDetail := details[constants.PurposeRevenueIwt]
			Expect(creditFeeDetail.Amount).To(decEqual(decimal10))
			Expect(creditFeeDetail.CurrencyCode).To(Equal("EUR"))
			Expect(creditFeeDetail.AccountId).To(BeNil())
			Expect(creditFeeDetail.RevenueAccountId).NotTo(BeNil())
			Expect(*creditFeeDetail.RevenueAccountId).To(BeEquivalentTo(456))
			Expect(creditFeeDetail.Transaction).NotTo(BeNil())

			creditFeeTx := creditFeeDetail.Transaction
			Expect(creditFeeTx.Id).To(BeNil(), "creditFeeTx should not be saved")
			Expect(*creditFeeTx.Amount).To(decEqual(str2Dec("10")))
			Expect(creditFeeTx.AccountId).To(BeNil())
			Expect(creditFeeTx.RevenueAccountId).NotTo(BeNil())
			Expect(*creditFeeTx.RevenueAccountId).To(BeEquivalentTo(456))
			Expect(creditFeeTx.CardId).To(BeNil())
			Expect(creditFeeTx.ShowAmount).To(BeNil())
			Expect(creditFeeTx.ShowAvailableBalanceSnapshot).To(BeNil())
			Expect(*creditFeeTx.AvailableBalanceSnapshot).To(decEqual(str2Dec("-90")))

			Expect(acc.Balance).To(decEqual(str2Dec("1090")))
			Expect(acc.AvailableAmount).To(decEqual(str2Dec("1090")))
			Expect(rev.Balance).To(decEqual(str2Dec("-90")))
			Expect(rev.AvailableAmount).To(decEqual(str2Dec("-90")))

		})

		It("should check Execute method", func() {
			acc := account("EUR", "1000")
			rev := revenueAccount("EUR", "0")

			feeParams := &fee.TransferFeeParams{
				Percent: str2Dec("10"),
			}

			input := NewCreditAccountInput(acc, true, true, rev, feeParams)
			mock.ExpectBegin()
			tx := gdb.Begin()
			unit := NewCreditAccount(tx, input, currencyBox)

			any := sqlmock.AnyArg()
			for i := 0; i < 4; i++ {
				mock.ExpectExec("INSERT INTO `transactions`.*").
					WithArgs(100, any, any, any, "executed", any, any, any, any, any, any, any, any, any, any, any, any).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}
			mock.ExpectExec("UPDATE `accounts`.*").
				WithArgs("1090", "1090", 999).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("UPDATE `revenue_accounts`.*").
				WithArgs("-90", "-90", 456).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("executed", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			_, err := unit.Execute(request("100", "EUR"))
			Expect(err).ShouldNot(HaveOccurred())

			err = mock.ExpectationsWereMet()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})

func request(amount, currencyCode string, referenceCode ...string) *requestModel.Request {
	a := str2Dec(amount)
	rqId := uint64(100)
	req := &requestModel.Request{
		Id:                    &rqId,
		Amount:                &a,
		BaseCurrencyCode:      &currencyCode,
		ReferenceCurrencyCode: &currencyCode,
		Status:                pointer.ToString("new"),
	}
	if referenceCode != nil {
		req.ReferenceCurrencyCode = &referenceCode[0]
	}
	return req
}

func revenueAccount(currencyCode, balance string) *model.RevenueAccountModel {
	b := str2Dec(balance)
	return &model.RevenueAccountModel{
		RevenueAccountPublic: model.RevenueAccountPublic{
			Balance:      b,
			CurrencyCode: currencyCode,
		},
		RevenueAccountPrivate: model.RevenueAccountPrivate{
			ID:              456,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			AvailableAmount: b,
		},
	}
}

func account(currencyCode, balance string) *model.Account {
	b := str2Dec(balance)
	return &model.Account{
		AccountPublic: model.AccountPublic{
			Number: currencyCode + "_MOCK_NUMBER",
			Type: &accountTypeModel.AccountType{
				AccountTypePublic: accountTypeModel.AccountTypePublic{
					CurrencyCode: currencyCode,
				},
			},
			TypeID:           123,
			UserId:           "fake-user-id",
			InitialBalance:   &b,
			AllowWithdrawals: pointer.ToBool(true),
			AllowDeposits:    pointer.ToBool(true),
		},
		AccountPrivate: model.AccountPrivate{
			ID:              999,
			AvailableAmount: b,
			Balance:         b,
		},
	}
}

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

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}
