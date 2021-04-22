package transfers_test

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	. "github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	mockTransfers "github.com/Confialink/wallet-accounts/internal/modules/request/transfers/mock"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	txModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/Confialink/wallet-pkg-utils/value"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var _ = Describe("Transfers", func() {
	var (
		mock                  sqlmock.Sqlmock
		gdb                   *gorm.DB
		sourceAccountEur      *model.Account
		sourceAccountUsd      *model.Account
		destinationAccountEur *model.Account
		destinationAccountUsd *model.Account
		revenueAccountEur     *model.RevenueAccountModel
		revenueAccountUsd     *model.RevenueAccountModel
	)
	_ = currencyBox.Add(euroCurrency)
	_ = currencyBox.Add(usdCurrency)
	Context("BA(Between Accounts) Transfer", func() {
		BeforeEach(func() {
			var db *sql.DB
			var err error
			sourceAccountEur = account("EUR", "1000")
			sourceAccountUsd = account("USD", "1000")
			destinationAccountEur = account("EUR", "0")
			destinationAccountUsd = account("USD", "0")
			revenueAccountEur = revenueAccount("EUR", "0")
			revenueAccountUsd = revenueAccount("USD", "0")

			db, mock, err = sqlmock.New() // mock sql.DB
			Expect(err).ShouldNot(HaveOccurred())

			gdb, err = gorm.Open("mysql", db) // open gorm db

			Expect(err).ShouldNot(HaveOccurred())
		})
		AfterEach(func() {
			err := mock.ExpectationsWereMet() // make sure all expectations were met
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should check bare transfer evaluation between 2 accounts in the same currency", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			input := NewBetweenAccountsInput(sourceAccountEur, destinationAccountEur, revenueAccountEur, str2Dec("0"), nil)
			unit := NewBetweenAccounts("ANY", currencyBox, input, nil, mockPF)

			details, err := unit.Evaluate(request("100", "EUR"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(2))
			Expect(details).To(HaveKey(constants.Purpose("any_outgoing")))
			Expect(details).To(HaveKey(constants.Purpose("any_incoming")))
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("900")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("900")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("0")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("0")))
			Expect(destinationAccountEur.Balance).To(decEqual(str2Dec("100")))
			Expect(destinationAccountEur.AvailableAmount).To(decEqual(str2Dec("100")))

			outgoingTx := details[constants.Purpose("any_outgoing")].Transaction
			Expect(outgoingTx.ShowAmount).To(BeNil())
		})
		It("should return error in case if request currency does not match given account currency", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			input := NewBetweenAccountsInput(sourceAccountEur, destinationAccountUsd, revenueAccountEur, str2Dec("0"), nil)
			unit := NewBetweenAccounts("ANY", currencyBox, input, nil, mockPF)

			_, err := unit.Evaluate(request("100", "EUR"))
			Expect(err).Should(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(transfer.ErrCurrenciesMismatch))

			input = NewBetweenAccountsInput(sourceAccountUsd, destinationAccountEur, revenueAccountEur, str2Dec("0"), nil)
			unit = NewBetweenAccounts("ANY", currencyBox, input, nil, mockPF)
			_, err = unit.Evaluate(request("100", "EUR"))
			Expect(err).Should(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(transfer.ErrCurrenciesMismatch))

			input = NewBetweenAccountsInput(sourceAccountEur, destinationAccountEur, revenueAccountUsd, str2Dec("0"), nil)
			unit = NewBetweenAccounts("ANY", currencyBox, input, nil, mockPF)
			_, err = unit.Evaluate(request("100", "EUR"))
			Expect(err).Should(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(transfer.ErrCurrenciesMismatch))
		})
		It("should evaluate transfer in different currencies", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()
			input := NewBetweenAccountsInput(
				sourceAccountEur,      // from this accounts
				destinationAccountUsd, // to this one
				revenueAccountEur,     // revenue does not expected, however account is required
				str2Dec("0"),          // no exchange margin
				nil,                   // no transfer fee
			)
			// 100 EUR -> to -> USD
			rqs := request("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("1.10")) // rate EUR/USD = 1.10

			unit := NewBetweenAccounts("ANY", currencyBox, input, nil, mockPF)

			details, err := unit.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(2))
			Expect(details).To(HaveKey(constants.Purpose("any_outgoing")))
			Expect(details).To(HaveKey(constants.Purpose("any_incoming")))
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("900")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("900")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("0")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("0")))
			Expect(destinationAccountUsd.Balance).To(decEqual(str2Dec("110")))
			Expect(destinationAccountUsd.AvailableAmount).To(decEqual(str2Dec("110")))

			outgoingTx := details[constants.Purpose("any_outgoing")].Transaction
			Expect(outgoingTx.ShowAmount).To(BeNil())
		})

		It("should evaluate transfer in different currencies with exchange margin fee", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			input := NewBetweenAccountsInput(
				sourceAccountEur,      // from this accounts
				destinationAccountUsd, // to this one
				revenueAccountEur,     // exchange margin fee must be credited to this revenue account
				str2Dec("10"),         // exchange margin is 10%
				nil,                   // no transfer fee
			)
			// 100 EUR -> to -> USD
			rqs := request("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("1.10")) // rate EUR/USD = 1.10

			unit := NewBetweenAccounts("ANY", currencyBox, input, nil, mockPF)

			details, err := unit.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(4))
			Expect(details).To(HaveKey(constants.Purpose("any_outgoing")))
			Expect(details).To(HaveKey(constants.Purpose("any_incoming")))
			Expect(details).To(HaveKey(constants.PurposeFeeExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeRevenueExchangeMargin))
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("900")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("900")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("10")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("10")))
			Expect(destinationAccountUsd.Balance).To(decEqual(str2Dec("99")))
			Expect(destinationAccountUsd.AvailableAmount).To(decEqual(str2Dec("99")))

			outgoingTx := details[constants.Purpose("any_outgoing")].Transaction
			Expect(outgoingTx.ShowAmount).ToNot(BeNil())
			Expect(*outgoingTx.ShowAmount).To(decEqual(str2Dec("-100")))

			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})
		It("should evaluate transfer in different currencies with exchange margin fee and transfer fee", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			feeParams := &fee.TransferFeeParams{
				Base:    str2Dec("10"), // Take 10 EUR
				Percent: str2Dec("25"), // +25%
			}
			input := NewBetweenAccountsInput(
				sourceAccountEur,      // from this accounts
				destinationAccountUsd, // to this one
				revenueAccountEur,     // exchange margin fee and transfer fee must be credited to this revenue account
				str2Dec("10"),         // exchange margin is 10%
				feeParams,             // no transfer fee
			)
			// 100 EUR -> to -> USD
			rqs := request("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("1.10")) // rate EUR/USD = 1.10

			unit := NewBetweenAccounts("ANY", currencyBox, input, nil, mockPF)

			details, err := unit.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(6))
			Expect(details).To(HaveKey(constants.Purpose("any_outgoing")))
			Expect(details).To(HaveKey(constants.Purpose("any_incoming")))
			Expect(details).To(HaveKey(constants.PurposeFeeExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeRevenueExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeFeeTransfer))
			Expect(details).To(HaveKey(constants.Purpose("revenue_any_transfer")))
			// 865 = 1000 - 100 outgoing - 35 transfer fee
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("865")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("865")))
			// 45 = 10 exchange margin + 35 transfer fee
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("45")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("45")))
			// 99 = ( 100 outgoing - 10 exchange margin ) * 1.10 rate
			Expect(destinationAccountUsd.Balance).To(decEqual(str2Dec("99")))
			Expect(destinationAccountUsd.AvailableAmount).To(decEqual(str2Dec("99")))

			type testData struct {
				purpose         string
				accId           bool
				cardId          bool
				revId           bool
				amount          decimal.Decimal
				visible         bool
				balanceSnapshot decimal.Decimal
				showAmount      *decimal.Decimal
			}

			testInput := []*testData{
				{"fee_exchange_margin", true, false, false, str2Dec("-10"), false, str2Dec("990"), nil},
				{"any_outgoing", true, false, false, str2Dec("-90"), true, str2Dec("900"), pointer.ToDecimal(str2Dec("-100"))},
				{"fee_default_transfer", true, false, false, str2Dec("-35"), true, str2Dec("865"), nil},
				{"any_incoming", true, false, false, str2Dec("99"), true, str2Dec("99"), nil},
				{"revenue_any_transfer", false, false, true, str2Dec("35"), true, str2Dec("35"), nil},
				{"revenue_exchange_margin", false, false, true, str2Dec("10"), true, str2Dec("45"), nil},
			}

			for _, data := range testInput {
				func(data *testData) {
					detail, ok := details[constants.Purpose(data.purpose)]
					Expect(ok).Should(BeTrue(), fmt.Sprintf("%s transaction must be presented in details", data.purpose))
					tx := detail.Transaction
					nilOrNot := map[bool]types.GomegaMatcher{
						true:  Not(BeNil()),
						false: BeNil(),
					}
					trueOrFalse := map[bool]types.GomegaMatcher{
						true:  BeEquivalentTo(true),
						false: BeEquivalentTo(false),
					}
					note := fmt.Sprintf("purpose %s", data.purpose)

					Expect(detail.AccountId).To(nilOrNot[data.accId], note)
					Expect(detail.CardId).To(nilOrNot[data.cardId], note)
					Expect(detail.RevenueAccountId).To(nilOrNot[data.revId], note)

					Expect(tx.AccountId).To(nilOrNot[data.accId], note)
					Expect(tx.CardId).To(nilOrNot[data.cardId], note)
					Expect(tx.RevenueAccountId).To(nilOrNot[data.revId], note)

					Expect(tx.IsVisible).ToNot(BeNil(), note)
					Expect(*tx.IsVisible).To(trueOrFalse[data.visible], note)

					Expect(*tx.Amount).To(decEqual(data.amount), note)
					Expect(*tx.AvailableBalanceSnapshot).To(decEqual(data.balanceSnapshot), note)
					if data.showAmount == nil {
						Expect(tx.ShowAmount).To(BeNil())
					} else {
						Expect(tx.ShowAmount).ToNot(BeNil())
						Expect(*tx.ShowAmount).To(decEqual(*data.showAmount))
					}
				}(data)
			}
			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})

		It(`should perform "pending" transfer between 2 accounts in the same currency`, func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			mockPermission := mockTransfers.NewMockPermissionChecker(ctrl)
			mockPermission.
				EXPECT().
				Check().
				Return(nil)
			mockPF.
				EXPECT().
				CreatePermission(gomock.Any(), gomock.Any()).
				Return(mockPermission, nil)

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
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("-100"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("900"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("1000"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "account",
					purpose:                      "any_outgoing",
				},
				{
					requestId:                    100,
					accountId:                    2,
					cardId:                       nil,
					revenueAccountId:             nil,
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("100"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("0"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("0"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "account",
					purpose:                      "any_incoming",
				},
			}

			for _, data := range testInput {
				func(data *testData) {
					mock.ExpectExec("INSERT INTO `transactions`.*").
						WithArgs(data.requestId, data.accountId, data.cardId, data.revenueAccountId, data.status, data.description, data.amount, data.showAmount, data.availableBalanceSnapshot, data.showAvailableBalanceSnapshot, data.isVisible, data.currentBalanceSnapshot, data.showCurrentBalanceSnapshot, data.type_, data.purpose, AnyTime{}, AnyTime{}).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}(data)
			}
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, account_id
				WithArgs(str2Dec("900"), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("pending", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			sourceAccountEur.ID = 1
			destinationAccountEur.ID = 2
			input := NewBetweenAccountsInput(sourceAccountEur, destinationAccountEur, revenueAccountEur, str2Dec("0"), nil)
			unit := NewBetweenAccounts("ANY", currencyBox, input, tx, mockPF)

			details, err := unit.Pending(request("100", "EUR"))
			Expect(details).To(HaveLen(2))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})

		It(`should perform "pending" transfer between 2 accounts in different currencies with margin and transfer fees`, func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			mockPermission := mockTransfers.NewMockPermissionChecker(ctrl)
			mockPermission.
				EXPECT().
				Check().
				Return(nil)
			mockPF.
				EXPECT().
				CreatePermission(gomock.Any(), gomock.Any()).
				Return(mockPermission, nil)

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
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("-10"),
					showAmount:                   nil,
					isVisible:                    false,
					availableBalanceSnapshot:     str2Dec("990"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("1000"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "fee",
					purpose:                      "fee_exchange_margin",
				},
				{
					requestId:                    100,
					accountId:                    1,
					cardId:                       nil,
					revenueAccountId:             nil,
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("-90"),
					showAmount:                   str2Dec("-100"),
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("900"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("1000"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "account",
					purpose:                      "any_outgoing",
				},
				{
					requestId:                    100,
					accountId:                    1,
					cardId:                       nil,
					revenueAccountId:             nil,
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("-35"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("865"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("1000"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "fee",
					purpose:                      "fee_default_transfer",
				},
				{
					requestId:                    100,
					accountId:                    2,
					cardId:                       nil,
					revenueAccountId:             nil,
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("99"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("0"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("0"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "account",
					purpose:                      "any_incoming",
				},
				{
					requestId:                    100,
					accountId:                    nil,
					cardId:                       nil,
					revenueAccountId:             3,
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("35"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("0"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("0"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "revenue",
					purpose:                      "revenue_any_transfer",
				},
				{
					requestId:                    100,
					accountId:                    nil,
					cardId:                       nil,
					revenueAccountId:             3,
					status:                       "pending",
					description:                  any,
					amount:                       str2Dec("10"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("0"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("0"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "revenue",
					purpose:                      "revenue_exchange_margin",
				},
			}

			for i, data := range testInput {
				func(data *testData) {
					mock.ExpectExec("INSERT INTO `transactions`.*").
						WithArgs(data.requestId, data.accountId, data.cardId, data.revenueAccountId, data.status, data.description, data.amount, data.showAmount, data.availableBalanceSnapshot, data.showAvailableBalanceSnapshot, data.isVisible, data.currentBalanceSnapshot, data.showCurrentBalanceSnapshot, data.type_, data.purpose, AnyTime{}, AnyTime{}).
						WillReturnResult(sqlmock.NewResult(int64(i), 1))
				}(data)
			}
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, account_id
				WithArgs(str2Dec("865"), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("pending", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			sourceAccountEur.ID = 1
			destinationAccountUsd.ID = 2
			revenueAccountEur.ID = 3
			feeParams := &fee.TransferFeeParams{
				Base:    str2Dec("10"), // Take 10 EUR
				Percent: str2Dec("25"), // +25%
			}
			input := NewBetweenAccountsInput(
				sourceAccountEur,      // from this accounts
				destinationAccountUsd, // to this one
				revenueAccountEur,     // exchange margin fee and transfer fee must be credited to this revenue account
				str2Dec("10"),         // exchange margin is 10%
				feeParams,             // no transfer fee
			)
			// 100 EUR -> to -> USD
			rqs := request("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("1.10")) // rate EUR/USD = 1.10

			unit := NewBetweenAccounts("ANY", currencyBox, input, tx, mockPF)

			details, err := unit.Pending(rqs)
			Expect(details).To(HaveLen(6))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})

		It(`should execute new transfer request between 2 accounts in the same currency`, func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			mockPermission := mockTransfers.NewMockPermissionChecker(ctrl)
			mockPermission.
				EXPECT().
				Check().
				Return(nil)
			mockPF.
				EXPECT().
				CreatePermission(gomock.Any(), gomock.Any()).
				Return(mockPermission, nil)

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
					amount:                       str2Dec("-100"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("900"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("900"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "account",
					purpose:                      "any_outgoing",
				},
				{
					requestId:                    100,
					accountId:                    2,
					cardId:                       nil,
					revenueAccountId:             nil,
					status:                       "executed",
					description:                  any,
					amount:                       str2Dec("100"),
					showAmount:                   nil,
					isVisible:                    true,
					availableBalanceSnapshot:     str2Dec("100"),
					showAvailableBalanceSnapshot: nil,
					currentBalanceSnapshot:       str2Dec("100"),
					showCurrentBalanceSnapshot:   nil,
					type_:                        "account",
					purpose:                      "any_incoming",
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
				WithArgs(str2Dec("900"), str2Dec("900"), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update destination account
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, balance, id
				WithArgs(str2Dec("100"), str2Dec("100"), 2).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update revenue account
			mock.ExpectExec("UPDATE `revenue_accounts`.*").
				// available_amount, balance, id
				WithArgs(str2Dec("0"), str2Dec("0"), 3).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("executed", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			sourceAccountEur.ID = 1
			destinationAccountEur.ID = 2
			revenueAccountEur.ID = 3
			input := NewBetweenAccountsInput(sourceAccountEur, destinationAccountEur, revenueAccountEur, str2Dec("0"), nil)
			unit := NewBetweenAccounts("ANY", currencyBox, input, tx, mockPF)

			details, err := unit.Execute(request("100", "EUR"))
			Expect(details).To(HaveLen(2))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})

		It(`should execute pending transfer request between 2 accounts in the same currency`, func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			mock.ExpectBegin()
			tx := gdb.Begin()

			accRowNames := []string{
				"id",
				"number",
				"type_id",
				"user_id",
				"initial_balance",
				"allow_withdrawals",
				"allow_deposit",
				"available_amount",
				"balance",
			}
			accTypeNames := []string{
				"id",
				"name",
				"currency_code",
			}
			sourceAccRows := sqlmock.NewRows(accRowNames)
			// balance permission should not impact existing pending requests
			// regarding of the fact that available amount is 99 less than requested amount 100
			sourceAccRows.AddRow(1, "EUR_1", 1, "user-uid", "1000", false, false, "99", "1000")

			destinationAccRows := sqlmock.NewRows(accRowNames)
			destinationAccRows.AddRow(2, "EUR_2", 1, "user-uid", "1000", false, false, "0", "0")

			revenueAccRows := sqlmock.NewRows([]string{
				"id",
				"currency_code",
				"balance",
				"available_amount",
			})
			revenueAccRows.AddRow(3, "EUR", "0", "0")

			txRows := sqlmock.NewRows([]string{
				"id",
				"request_id",
				"account_id",
				"card_id",
				"revenue_account_id",
				"status",
				"amount",
				"show_amount",
				"show_balance_snapshot",
				"is_visible",
				"balance_snapshot",
				"type",
				"purpose",
			})
			txRows.AddRow(1, 100, 1, nil, nil, "pending", "-100", nil, nil, true, "900", "account", "any_outgoing")
			txRows.AddRow(2, 100, 2, nil, nil, "pending", "100", nil, nil, true, "100", "account", "any_incoming")

			mock.
				ExpectQuery("^SELECT (.+) FROM `transactions` WHERE (.+)").
				// request_id
				WithArgs(100).
				WillReturnRows(txRows)

			// select source account
			mock.
				ExpectQuery("SELECT (.+) FROM `accounts` WHERE `accounts`.`id` .* FOR UPDATE").
				WithArgs(1).
				WillReturnRows(sourceAccRows)
			accountTypeRows := sqlmock.NewRows(accTypeNames)
			accountTypeRows.AddRow(1, "EUR_TYPE", "EUR")
			mock.ExpectQuery("SELECT (.+) FROM `account_types`.*").
				WillReturnRows(accountTypeRows)
			// select destination account
			mock.
				ExpectQuery("SELECT (.+) FROM `accounts` WHERE `accounts`.`id` .* FOR UPDATE").
				WithArgs(2).
				WillReturnRows(destinationAccRows)
			accountTypeRows = sqlmock.NewRows(accTypeNames)
			accountTypeRows.AddRow(1, "EUR_TYPE", "EUR")
			mock.ExpectQuery("SELECT (.+) FROM `account_types`.*").
				WillReturnRows(accountTypeRows)
			// select revenue account
			mock.ExpectQuery("SELECT (.+) FROM `revenue_accounts`.*").
				WillReturnRows(revenueAccRows)

			type testData struct {
				id                       interface{}
				status                   interface{}
				amount                   interface{}
				showAmount               interface{}
				availableBalanceSnapshot interface{}
				currentBalanceSnapshot   interface{}
			}

			testInput := []*testData{
				{
					id:                       1,
					status:                   "executed",
					amount:                   "-100",
					showAmount:               nil,
					availableBalanceSnapshot: "99",
					currentBalanceSnapshot:   "900",
				},
				{
					id:                       2,
					status:                   "executed",
					amount:                   "100",
					showAmount:               nil,
					availableBalanceSnapshot: "100",
					currentBalanceSnapshot:   "100",
				},
			}

			for _, data := range testInput {
				func(data *testData) {
					mock.ExpectExec("UPDATE `transactions`.*").
						WithArgs(data.amount, data.availableBalanceSnapshot, data.currentBalanceSnapshot, data.showAmount, data.status, AnyTime{}, data.id).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}(data)
			}

			// update source account
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, balance, id
				WithArgs("99", "900", 1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update destination account
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, balance, id
				WithArgs("100", "100", 2).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update revenue account
			mock.ExpectExec("UPDATE `revenue_accounts`.*").
				// available_amount, balance, id
				WithArgs("0", "0", 3).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("executed", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			rqs := request("100", "EUR")
			rqs.Status = pointer.ToString("pending")
			rqs.GetInput().Set("sourceAccountId", 1)
			rqs.GetInput().Set("destinationAccountId", 2)
			rqs.GetInput().Set("revenueAccountId", 3)
			rqs.GetInput().Set("exchangeMarginPercent", "0")
			rqs.GetInput().Set("transferFeeParams", nil)

			input := NewDbBetweenAccountsInput(tx, rqs, nil)
			unit := NewBetweenAccounts("ANY", currencyBox, input, tx, mockPF)

			details, err := unit.Execute(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(2))
			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})

		It("should cancel pending transfer request", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			mock.ExpectBegin()
			tx := gdb.Begin()

			rows := sqlmock.NewRows([]string{
				"id",
				"request_id",
				"account_id",
				"card_id",
				"revenue_account_id",
				"status",
				"amount",
				"show_amount",
				"show_balance_snapshot",
				"is_visible",
				"balance_snapshot",
				"type",
				"purpose",
			})
			rows.AddRow(1, 100, 1, nil, nil, "pending", "-10", nil, nil, true, "990", "fee", "any_outgoing")
			rows.AddRow(2, 100, 1, nil, nil, "pending", "-90", nil, nil, true, "900", "account", "any_outgoing")
			rows.AddRow(3, 100, 2, nil, nil, "pending", "99", nil, nil, true, "0", "account", "any_incoming")
			rows.AddRow(4, 100, nil, nil, 3, "pending", "99", nil, nil, true, "0", "revenue", "any_incoming")

			mock.
				ExpectQuery("^SELECT (.+) FROM `transactions` WHERE (.+)").
				// request_id
				WithArgs(100).
				WillReturnRows(rows)

			sourceAccountEur.ID = 1
			sourceAccountEur.AvailableAmount = str2Dec("900") // simulate account in pending state

			input := NewBetweenAccountsInput(sourceAccountEur, destinationAccountUsd, revenueAccountEur, str2Dec("10"), nil)
			unit := NewBetweenAccounts("ANY", currencyBox, input, tx, mockPF)

			rqs := request("100", "EUR", "USD")
			rqs.Status = pointer.ToString("pending")
			rqs.Rate = pointer.ToDecimal(str2Dec("1.10")) // rate EUR/USD = 1.10
			// update transactions statuses
			mock.
				ExpectExec("UPDATE `transactions`.*").
				// status, request_id
				WithArgs("cancelled", 100).
				WillReturnResult(sqlmock.NewResult(1, 2))
			// update source account available amount
			mock.
				ExpectExec("UPDATE `accounts`.*").
				// status, request_id
				WithArgs("1000", "1000", 1).
				WillReturnResult(sqlmock.NewResult(1, 2))

			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("cancelled", "test", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := unit.Cancel(rqs, "test")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("1000")))
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("1000")))
		})

		It("should modify pending transfer between 2 accounts", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			mock.ExpectBegin()
			tx := gdb.Begin()

			rows := sqlmock.NewRows([]string{
				"id",
				"request_id",
				"account_id",
				"card_id",
				"revenue_account_id",
				"status",
				"amount",
				"show_amount",
				"is_visible",
				"balance_snapshot",
				"type",
				"purpose",
			})
			rows.AddRow(1, 100, 1, nil, nil, "pending", "-100", nil, true, "900", "account", "any_outgoing")
			rows.AddRow(2, 100, 2, nil, nil, "pending", "110", nil, true, "0", "account", "any_incoming")

			mock.
				ExpectQuery("^SELECT (.+) FROM `transactions` WHERE (.+)").
				// request_id
				WithArgs(100).
				WillReturnRows(rows)

			sourceAccountEur.ID = 1
			sourceAccountEur.AvailableAmount = str2Dec("900")

			destinationAccountUsd.ID = 2

			input := NewBetweenAccountsInput(
				sourceAccountEur,
				destinationAccountUsd,
				revenueAccountEur,
				str2Dec("0"),
				nil,
			)
			unit := NewBetweenAccounts("ANY", currencyBox, input, tx, mockPF)

			rqs := request("100", "EUR", "USD")
			// simulate the case where old rate is 1.1, updated rate is 1.2
			rqs.Rate = pointer.ToDecimal(str2Dec("1.2"))

			mock.
				ExpectExec("UPDATE `transactions`.*").
				// amount, available_balance_snapshot, current_balance_snapshot, show_amount, id
				WithArgs("-100", "900", "1000", nil, "pending", AnyTime{}, 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.
				ExpectExec("UPDATE `transactions`.*").
				// amount, available_balance_snapshot, current_balance_snapshot, show_amount, id
				WithArgs("120", "0", "0", nil, "pending", AnyTime{}, 2).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// update request
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("100", "1.2", 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			_, err := unit.Modify(rqs)
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrUnexpectedStatus))

			rqs.Status = pointer.ToString("pending")
			details, err := unit.Modify(rqs)
			Expect(err).ToNot(HaveOccurred())
			Expect(details).To(HaveLen(2))
			Expect(ensureTransactionsOrder(unit.Transactions())).To(Succeed())
		})
	})
})

func ensureTransactionsOrder(transactions []*txModel.Transaction) error {
	visible := make(map[string]map[int64]*txModel.Transaction)
	for _, tx := range transactions {
		txType, id := txTypeAndId(tx)
		if !*tx.IsVisible {
			if causeTx, exist := visible[txType][id]; exist {
				note := fmt.Sprintf(`"type:%s, id: %d"`, txType, id)
				return errors.New(fmt.Sprintf(
					"[%s] transaction %s (%s) is invisible but it comes after visible transaction %s (%s), all invisible transactions must be created before any visible transaction related to the same balance %s",
					note,
					value.FromString(causeTx.Purpose),
					value.FromString(causeTx.Description),
					value.FromString(tx.Purpose),
					value.FromString(tx.Description),
					note,
				))
			}
			continue
		}
		if _, ok := visible[txType]; !ok {
			visible[txType] = make(map[int64]*txModel.Transaction)
		}
		visible[txType][id] = tx
	}
	return nil
}

func txTypeAndId(transaction *txModel.Transaction) (txType string, id int64) {
	if transaction.AccountId != nil {
		txType = "account"
		id = int64(*transaction.AccountId)
		return
	}
	if transaction.RevenueAccountId != nil {
		txType = "revenue"
		id = int64(*transaction.RevenueAccountId)
		return
	}
	if transaction.CardId != nil {
		txType = "card"
		id = int64(*transaction.CardId)
		return
	}
	txType = "unknown"
	id = 0
	return
}
