package transfers_test

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-accounts/internal/transfer"
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
		revenueAccountEur *model.RevenueAccountModel
	)
	_ = currencyBox.Add(euroCurrency)
	_ = currencyBox.Add(usdCurrency)
	Context("Deduct Revenue Account(DRA)", func() {
		BeforeEach(func() {
			var db *sql.DB
			var err error

			revenueAccountEur = revenueAccount("EUR", "1000")
			db, mock, err = sqlmock.New() // mock sql.DB
			Expect(err).ShouldNot(HaveOccurred())

			gdb, err = gorm.Open("mysql", db) // open gorm db

			Expect(err).ShouldNot(HaveOccurred())
		})
		AfterEach(func() {
			err := mock.ExpectationsWereMet() // make sure all expectations were met
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should evaluate deduct revenue account", func() {
			input := transfers.NewDraInput(revenueAccountEur)
			dra := transfers.NewDeductRevenueAccount(gdb, input, currencyBox)

			details, err := dra.Evaluate(request("1100", "EUR"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(1))
			Expect(details).To(HaveKey(constants.PurposeDebitRevenue))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("-100")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("-100")))
		})

		When("revenue and request currency do not match", func() {
			It("should raise error", func() {
				input := transfers.NewDraInput(revenueAccountEur)
				dra := transfers.NewDeductRevenueAccount(gdb, input, currencyBox)

				_, err := dra.Evaluate(request("100", "CHF"))
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(transfer.ErrCurrenciesMismatch))

				_, err = dra.Execute(request("100", "CHF"))
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(transfer.ErrCurrenciesMismatch))
			})
		})
		When("request has wrong status", func() {
			It("should not execute", func() {
				input := transfers.NewDraInput(revenueAccountEur)
				dra := transfers.NewDeductRevenueAccount(gdb, input, currencyBox)
				rqs := request("100", "EUR")
				unexpectedStatuses := []string{
					"pending",
					"cancelled",
					"executed",
				}
				rqs.Status = nil
				_, err := dra.Execute(rqs)
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(transfers.ErrMissingRequestData))

				for _, s := range unexpectedStatuses {
					func(s string) {
						rqs.Status = &s
						_, err := dra.Execute(rqs)
						Expect(err).Should(HaveOccurred())
						Expect(errors.Cause(err)).To(Equal(transfers.ErrUnexpectedStatus))
					}(s)
				}
			})
		})

		It("should execute dra request", func() {
			mock.ExpectBegin()
			tx := gdb.Begin()

			revenueAccountEur.ID = 3
			input := transfers.NewDraInput(revenueAccountEur)
			dra := transfers.NewDeductRevenueAccount(tx, input, currencyBox)

			any := sqlmock.AnyArg()

			mock.
				ExpectExec("INSERT INTO `transactions`.*").
				//request_id, account_id, card_id, revenue_account_id, status, description, amount, show_amount, available_balance_snapshot, show_available_balance_snapshot, is_visible, current_balance_snapshot, show_current_balance_snapshot, type, purpose, created_at, updated_at
				WithArgs(100, nil, nil, 3, "executed", any, "-100", nil, "900", nil, true, "900", nil, "revenue", "debit_revenue", AnyTime{}, AnyTime{}).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.
				ExpectExec("UPDATE `revenue_accounts`.*").
				// available_amount, balance, id
				WithArgs("900", "900", 3).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("executed", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			details, err := dra.Execute(request("100", "EUR"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(1))
			Expect(details).To(HaveKey(constants.PurposeDebitRevenue))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("900")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("900")))
		})
	})
})
