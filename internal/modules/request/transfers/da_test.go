package transfers_test

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("Transfers", func() {
	var (
		mock              sqlmock.Sqlmock
		gdb               *gorm.DB
		sourceAccountEur  *model.Account
		revenueAccountEur *model.RevenueAccountModel
	)
	_ = currencyBox.Add(euroCurrency)
	_ = currencyBox.Add(usdCurrency)
	Context("DA(Debit Account) Transfer", func() {
		BeforeEach(func() {
			var db *sql.DB
			var err error
			sourceAccountEur = account("EUR", "1000")
			revenueAccountEur = revenueAccount("EUR", "0")

			db, mock, err = sqlmock.New() // mock sql.DB
			Expect(err).ShouldNot(HaveOccurred())

			gdb, err = gorm.Open("mysql", db) // open gorm db

			Expect(err).ShouldNot(HaveOccurred())
		})
		AfterEach(func() {
			err := mock.ExpectationsWereMet() // make sure all expectations were met
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should evaluate simple debit account", func() {
			input := transfers.NewDaInput(sourceAccountEur, nil, false, false)
			da := transfers.NewDebitAccount(nil, input, currencyBox)

			rqs := request("100", "EUR")
			details, err := da.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(1))
			Expect(details).To(HaveKey(constants.PurposeDebitAccount))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("900")))
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("900")))
		})

		It("should evaluate debit account with credit to revenue", func() {
			input := transfers.NewDaInput(sourceAccountEur, revenueAccountEur, true, false)
			da := transfers.NewDebitAccount(nil, input, currencyBox)

			rqs := request("100", "EUR")
			details, err := da.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(2))
			Expect(details).To(HaveKey(constants.PurposeDebitAccount))
			Expect(details).To(HaveKey(constants.PurposeCreditRevenue))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("900")))
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("900")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("100")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("100")))
		})

		It("should ensure that allowNegativeAmount permission is respected", func() {
			input := transfers.NewDaInput(sourceAccountEur, revenueAccountEur, true, false)
			unit := transfers.NewDebitAccount(gdb, input, currencyBox)

			_, err := unit.Execute(request("1001", "EUR"))
			Expect(err).Should(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(transfers.ErrInsufficientBalance))
		})

		It("execute new debit account transfer", func() {
			mock.ExpectBegin()
			tx := gdb.Begin()
			any := sqlmock.AnyArg()
			type testData struct {
				requestId                    interface{}
				accountId                    interface{}
				cardId                       interface{}
				revenueAccountId             interface{}
				status                       interface{}
				description                  interface{}
				amount                       interface{}
				showAmount                   interface{}
				availableBalanceSnapshot     interface{}
				showAvailableBalanceSnapshot interface{}
				currentBalanceSnapshot       interface{}
				showCurrentBalanceSnapshot   interface{}
				isVisible                    interface{}
				type_                        interface{}
				purpose                      interface{}
			}

			testInput := []*testData{
				{
					requestId:                    100,
					accountId:                    1,
					cardId:                       nil,
					revenueAccountId:             nil,
					status:                       "executed",
					description:                  any,
					amount:                       str2Dec("-1001"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("-1"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("-1"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "account",
					purpose:                      "debit_account",
				},
				{
					requestId:                    100,
					accountId:                    nil,
					cardId:                       nil,
					revenueAccountId:             3,
					status:                       "executed",
					description:                  any,
					amount:                       str2Dec("1001"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("1001"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("1001"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "revenue",
					purpose:                      "credit_revenue",
				},
			}

			for _, data := range testInput {
				func(data *testData) {
					mock.ExpectExec("INSERT INTO `transactions`.*").
						WithArgs(data.requestId, data.accountId, data.cardId, data.revenueAccountId, data.status, data.description, data.amount, data.showAmount, data.availableBalanceSnapshot, data.showAvailableBalanceSnapshot, data.isVisible, data.currentBalanceSnapshot, data.showCurrentBalanceSnapshot, data.type_, data.purpose, AnyTime{}, AnyTime{}).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}(data)
			}
			// update source account
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, balance, id
				WithArgs(str2Dec("-1"), str2Dec("-1"), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// update revenue account
			mock.ExpectExec("UPDATE `revenue_accounts`.*").
				// available_amount, balance, id
				WithArgs(str2Dec("1001"), str2Dec("1001"), 3).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("executed", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			sourceAccountEur.ID = 1
			revenueAccountEur.ID = 3
			input := transfers.NewDaInput(sourceAccountEur, revenueAccountEur, true, true)
			unit := transfers.NewDebitAccount(tx, input, currencyBox)

			details, err := unit.Execute(request("1001", "EUR"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(2))
			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})
	})
})
