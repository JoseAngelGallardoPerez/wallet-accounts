package transfers_test

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	. "github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	mockTransfers "github.com/Confialink/wallet-accounts/internal/modules/request/transfers/mock"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transfers", func() {
	var (
		mock             sqlmock.Sqlmock
		gdb              *gorm.DB
		sourceAccountEur *model.Account
		//sourceAccountUsd      *model.Account
		//destinationAccountEur *model.Account
		//destinationAccountUsd *model.Account
		revenueAccountEur *model.RevenueAccountModel
		//revenueAccountUsd     *model.RevenueAccountModel
	)
	_ = currencyBox.Add(euroCurrency)
	_ = currencyBox.Add(usdCurrency)
	owtRequest := func(amount, currencyCode string, referenceCode ...string) *requestModel.Request {
		rqs := request(amount, currencyCode, referenceCode...)
		rqs.RateDesignation = requestModel.RateDesignationReferenceBase
		rqs.InputAmount = rqs.Amount
		rqs.Amount = nil
		rqs.Rate = pointer.ToDecimal(str2Dec("1"))
		return rqs
	}

	Context("OWT(Outgoing Wire Transfer)", func() {
		BeforeEach(func() {
			var db *sql.DB
			var err error
			sourceAccountEur = account("EUR", "1000")
			//sourceAccountUsd = account("USD", "1000")
			//destinationAccountEur = account("EUR", "0")
			//destinationAccountUsd = account("USD", "0")
			revenueAccountEur = revenueAccount("EUR", "0")
			//revenueAccountUsd = revenueAccount("USD", "0")

			db, mock, err = sqlmock.New() // mock sql.DB
			Expect(err).ShouldNot(HaveOccurred())

			gdb, err = gorm.Open("mysql", db) // open gorm db

			Expect(err).ShouldNot(HaveOccurred())
		})
		AfterEach(func() {
			err := mock.ExpectationsWereMet() // make sure all expectations were met
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should check bare transfer evaluation in the same currency", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			input := NewOwtInput(
				sourceAccountEur,
				revenueAccountEur,
				str2Dec("0"),
				nil,
				"",
				"",
			)
			rqs := owtRequest("100", "EUR")
			owt := NewOutgoingWireTransfer(input, currencyBox, gdb, mockPF)

			details, err := owt.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(1))
			Expect(details).To(HaveKey(constants.Purpose("owt_outgoing")))
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("900")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("900")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("0")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("0")))
		})

		It("should check transfer evaluation in different currencies with margin fee", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			input := NewOwtInput(
				sourceAccountEur,
				revenueAccountEur,
				str2Dec("10"),
				nil,
				"",
				"",
			)
			rqs := owtRequest("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("0.9"))
			owt := NewOutgoingWireTransfer(input, currencyBox, gdb, mockPF)

			details, err := owt.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(3))
			Expect(details).To(HaveKey(constants.PurposeOWTOutgoing))
			Expect(details).To(HaveKey(constants.PurposeFeeExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeRevenueExchangeMargin))
			// 1000 - (100 * 0.9)outgoing - (100 * 0.9 * 0.1)margin
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("901")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("901")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("9")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("9")))
		})

		It("should check transfer evaluation in different currencies with margin and transfer fees", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				Return(mockPF).
				AnyTimes()

			feeParams := &fee.TransferFeeParams{Base: str2Dec("10")}
			input := NewOwtInput(
				sourceAccountEur,
				revenueAccountEur,
				str2Dec("10"),
				feeParams,
				"",
				"",
			)
			rqs := owtRequest("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("0.9"))
			owt := NewOutgoingWireTransfer(input, currencyBox, gdb, mockPF)

			details, err := owt.Evaluate(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(5))
			Expect(details).To(HaveKey(constants.PurposeOWTOutgoing))
			Expect(details).To(HaveKey(constants.PurposeFeeExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeRevenueExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeFeeTransfer))
			Expect(details).To(HaveKey(constants.Purpose("revenue_owt_transfer")))
			// 1000 - (100 * 0.9)outgoing - (100 * 0.9 * 0.1)margin - (10)transfer fee
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("891")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("891")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("19")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("19")))
		})

		It("should make pending transfer request", func() {
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
				requestId                interface{}
				accountId                interface{}
				cardId                   interface{}
				revenueAccountId         interface{}
				status                   interface{}
				description              interface{}
				amount                   interface{}
				showAmount               interface{}
				availableBalanceSnapshot interface{}
				currentBalanceSnapshot   interface{}
				isVisible                interface{}
				type_                    interface{}
				purpose                  interface{}
			}

			testInput := []*testData{
				{
					requestId:                100,
					accountId:                1,
					cardId:                   nil,
					revenueAccountId:         nil,
					status:                   "pending",
					description:              any,
					amount:                   "-9",
					showAmount:               nil,
					availableBalanceSnapshot: "991",
					currentBalanceSnapshot:   "1000",
					isVisible:                false,
					type_:                    "fee",
					purpose:                  "fee_exchange_margin",
				},
				{
					requestId:                100,
					accountId:                1,
					cardId:                   nil,
					revenueAccountId:         nil,
					status:                   "pending",
					description:              any,
					amount:                   "-90",
					showAmount:               "-99",
					isVisible:                true,
					availableBalanceSnapshot: "901",
					currentBalanceSnapshot:   "1000",
					type_:                    "account",
					purpose:                  "owt_outgoing",
				},
				{
					requestId:                100,
					accountId:                1,
					cardId:                   nil,
					revenueAccountId:         nil,
					status:                   "pending",
					description:              any,
					amount:                   str2Dec("-10"),
					showAmount:               nil,
					isVisible:                true,
					availableBalanceSnapshot: "891",
					currentBalanceSnapshot:   "1000",
					type_:                    "fee",
					purpose:                  "fee_default_transfer",
				},
				{
					requestId:                100,
					accountId:                nil,
					cardId:                   nil,
					revenueAccountId:         3,
					status:                   "pending",
					description:              any,
					amount:                   "10",
					showAmount:               nil,
					availableBalanceSnapshot: "0",
					currentBalanceSnapshot:   "0",
					isVisible:                true,
					type_:                    "revenue",
					purpose:                  "revenue_owt_transfer",
				},
				{
					requestId:                100,
					accountId:                nil,
					cardId:                   nil,
					revenueAccountId:         3,
					status:                   "pending",
					description:              any,
					amount:                   "9",
					showAmount:               nil,
					availableBalanceSnapshot: "0",
					currentBalanceSnapshot:   "0",
					isVisible:                true,
					type_:                    "revenue",
					purpose:                  "revenue_exchange_margin",
				},
			}

			for _, data := range testInput {
				func(data *testData) {
					mock.ExpectExec("INSERT INTO `transactions`.*").
						WithArgs(data.requestId, data.accountId, data.cardId, data.revenueAccountId, data.status, data.description, data.amount, data.showAmount, data.availableBalanceSnapshot, any, data.isVisible, data.currentBalanceSnapshot, any, data.type_, data.purpose, AnyTime{}, AnyTime{}).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}(data)
			}
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, account_id
				WithArgs("891", "1000", 1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("pending", "90", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			sourceAccountEur.ID = 1
			revenueAccountEur.ID = 3

			feeParams := &fee.TransferFeeParams{Base: str2Dec("10")}
			input := NewOwtInput(
				sourceAccountEur,
				revenueAccountEur,
				str2Dec("10"),
				feeParams,
				"",
				"",
			)
			rqs := owtRequest("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("0.9"))
			owt := NewOutgoingWireTransfer(input, currencyBox, tx, mockPF)

			details, err := owt.Pending(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(5))
			Expect(details).To(HaveKey(constants.PurposeOWTOutgoing))
			Expect(details).To(HaveKey(constants.PurposeFeeExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeRevenueExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeFeeTransfer))
			Expect(details).To(HaveKey(constants.Purpose("revenue_owt_transfer")))
			// 1000 - (100 * 0.9)outgoing - (100 * 0.9 * 0.1)margin - (10)transfer fee
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("1000")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("891")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("0")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("0")))
			Expect(ensureTransactionsOrder(owt.Transactions())).To(Succeed())

		})

		It("should execute new transfer request", func() {
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
				requestId                interface{}
				accountId                interface{}
				cardId                   interface{}
				revenueAccountId         interface{}
				status                   interface{}
				description              interface{}
				amount                   interface{}
				showAmount               interface{}
				availableBalanceSnapshot interface{}
				currentBalanceSnapshot   interface{}
				isVisible                interface{}
				type_                    interface{}
				purpose                  interface{}
			}

			testInput := []*testData{
				{
					requestId:                100,
					accountId:                1,
					cardId:                   nil,
					revenueAccountId:         nil,
					status:                   "executed",
					description:              any,
					amount:                   "-9",
					showAmount:               nil,
					availableBalanceSnapshot: "991",
					currentBalanceSnapshot:   "991",
					isVisible:                false,
					type_:                    "fee",
					purpose:                  "fee_exchange_margin",
				},
				{
					requestId:                100,
					accountId:                1,
					cardId:                   nil,
					revenueAccountId:         nil,
					status:                   "executed",
					description:              any,
					amount:                   "-90",
					showAmount:               "-99",
					isVisible:                true,
					availableBalanceSnapshot: "901",
					currentBalanceSnapshot:   "901",
					type_:                    "account",
					purpose:                  "owt_outgoing",
				},
				{
					requestId:                100,
					accountId:                1,
					cardId:                   nil,
					revenueAccountId:         nil,
					status:                   "executed",
					description:              any,
					amount:                   str2Dec("-10"),
					showAmount:               nil,
					isVisible:                true,
					availableBalanceSnapshot: "891",
					currentBalanceSnapshot:   "891",
					type_:                    "fee",
					purpose:                  "fee_default_transfer",
				},
				{
					requestId:                100,
					accountId:                nil,
					cardId:                   nil,
					revenueAccountId:         3,
					status:                   "executed",
					description:              any,
					amount:                   "10",
					showAmount:               nil,
					availableBalanceSnapshot: "10",
					currentBalanceSnapshot:   "10",
					isVisible:                true,
					type_:                    "revenue",
					purpose:                  "revenue_owt_transfer",
				},
				{
					requestId:                100,
					accountId:                nil,
					cardId:                   nil,
					revenueAccountId:         3,
					status:                   "executed",
					description:              any,
					amount:                   "9",
					showAmount:               nil,
					availableBalanceSnapshot: "19",
					currentBalanceSnapshot:   "19",
					isVisible:                true,
					type_:                    "revenue",
					purpose:                  "revenue_exchange_margin",
				},
			}

			for _, data := range testInput {
				func(data *testData) {
					mock.ExpectExec("INSERT INTO `transactions`.*").
						WithArgs(data.requestId, data.accountId, data.cardId, data.revenueAccountId, data.status, data.description, data.amount, data.showAmount, data.availableBalanceSnapshot, any, data.isVisible, data.currentBalanceSnapshot, any, data.type_, data.purpose, AnyTime{}, AnyTime{}).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}(data)
			}
			mock.ExpectExec("UPDATE `accounts`.*").
				WithArgs("891", "891", 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// update revenue account
			mock.ExpectExec("UPDATE `revenue_accounts`.*").
				// available_amount, balance, id
				WithArgs("19", "19", 3).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("executed", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			sourceAccountEur.ID = 1
			revenueAccountEur.ID = 3

			feeParams := &fee.TransferFeeParams{Base: str2Dec("10")}
			input := NewOwtInput(
				sourceAccountEur,
				revenueAccountEur,
				str2Dec("10"),
				feeParams,
				"",
				"",
			)
			rqs := owtRequest("100", "EUR", "USD")
			rqs.Rate = pointer.ToDecimal(str2Dec("0.9"))
			owt := NewOutgoingWireTransfer(input, currencyBox, tx, mockPF)

			details, err := owt.Execute(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(5))
			Expect(details).To(HaveKey(constants.PurposeOWTOutgoing))
			Expect(details).To(HaveKey(constants.PurposeFeeExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeRevenueExchangeMargin))
			Expect(details).To(HaveKey(constants.PurposeFeeTransfer))
			Expect(details).To(HaveKey(constants.Purpose("revenue_owt_transfer")))
			// 1000 - (100 * 0.9)outgoing - (100 * 0.9 * 0.1)margin - (10)transfer fee
			Expect(sourceAccountEur.Balance).To(decEqual(str2Dec("891")))
			Expect(sourceAccountEur.AvailableAmount).To(decEqual(str2Dec("891")))
			Expect(revenueAccountEur.Balance).To(decEqual(str2Dec("19")))
			Expect(revenueAccountEur.AvailableAmount).To(decEqual(str2Dec("19")))
			Expect(ensureTransactionsOrder(owt.Transactions())).To(Succeed())
		})

		It(`should execute pending transfer request`, func() {
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
			// regarding of the fact that available amount is 50 less than requested amount 100
			sourceAccRows.AddRow(1, "EUR_1", 1, "user-uid", "1000", false, false, "50", "1000")

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
				"is_visible",
				"balance_snapshot",
				"type",
				"purpose",
			})
			txRows.AddRow(1, 100, 1, nil, nil, "pending", "-9", nil, false, "900", "account", "fee_exchange_margin")
			txRows.AddRow(2, 100, 1, nil, nil, "pending", "-90", nil, true, "900", "account", "owt_outgoing")
			txRows.AddRow(3, 100, nil, nil, 3, "pending", "9", nil, true, "100", "account", "revenue_exchange_margin")

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
					amount:                   "-9",
					showAmount:               nil,
					availableBalanceSnapshot: "50",
					currentBalanceSnapshot:   "991",
				},
				{
					id:                       2,
					status:                   "executed",
					amount:                   "-90",
					showAmount:               "-99",
					availableBalanceSnapshot: "50",
					currentBalanceSnapshot:   "901",
				},
				{
					id:                       3,
					status:                   "executed",
					amount:                   "9",
					showAmount:               nil,
					availableBalanceSnapshot: "9",
					currentBalanceSnapshot:   "9",
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
				WithArgs(str2Dec("50"), str2Dec("901"), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update revenue account
			mock.ExpectExec("UPDATE `revenue_accounts`.*").
				// available_amount, balance, id
				WithArgs(str2Dec("9"), str2Dec("9"), 3).
				WillReturnResult(sqlmock.NewResult(1, 1))
			// update request status
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("executed", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			rqs := owtRequest("100", "EUR", "USD")
			rqs.Status = pointer.ToString("pending")
			rqs.Rate = pointer.ToDecimal(str2Dec("0.9"))
			rqs.GetInput().Set("sourceAccountId", 1)
			rqs.GetInput().Set("revenueAccountId", 3)
			rqs.GetInput().Set("exchangeMarginPercent", "10")
			rqs.GetInput().Set("transferFeeParams", nil)
			rqs.GetInput().Set("refMessage", "")
			rqs.GetInput().Set("beneficiaryCustomerAccountName", "")

			input := NewDbOWTInput(tx, rqs, nil)
			unit := NewOutgoingWireTransfer(input, currencyBox, tx, mockPF)

			details, err := unit.Execute(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(details).To(HaveLen(3))
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
			// regarding of the fact that available amount is 50 less than requested amount 100
			sourceAccRows.AddRow(1, "EUR_1", 1, "user-uid", "1000", false, false, "50", "1000")

			txRows := sqlmock.NewRows([]string{
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
			txRows.AddRow(1, 100, 1, nil, nil, "pending", "-9", nil, false, "900", "account", "fee_exchange_margin")
			txRows.AddRow(2, 100, 1, nil, nil, "pending", "-90", nil, true, "900", "account", "owt_outgoing")
			txRows.AddRow(3, 100, nil, nil, 3, "pending", "9", nil, true, "100", "account", "revenue_exchange_margin")

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

			mock.ExpectExec("UPDATE `transactions`.*").
				WithArgs("cancelled", 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// update source account
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, balance, id
				WithArgs(str2Dec("149"), str2Dec("1000"), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// update request status and cancellation reason
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("cancelled", "test", AnyTime{}, 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			rqs := owtRequest("100", "EUR", "USD")
			rqs.Status = pointer.ToString("pending")
			rqs.Rate = pointer.ToDecimal(str2Dec("0.9"))
			rqs.GetInput().Set("sourceAccountId", 1)
			rqs.GetInput().Set("revenueAccountId", 3)
			rqs.GetInput().Set("exchangeMarginPercent", "10")
			rqs.GetInput().Set("transferFeeParams", nil)
			rqs.GetInput().Set("refMessage", "")
			rqs.GetInput().Set("beneficiaryCustomerAccountName", "")

			input := NewDbOWTInput(tx, rqs, nil)
			unit := NewOutgoingWireTransfer(input, currencyBox, tx, mockPF)

			err := unit.Cancel(rqs, "test")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should modify pending transaction", func() {
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
			sourceAccRows.AddRow(1, "EUR_1", 1, "user-uid", "1000", true, true, "901", "1000")

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
				"is_visible",
				"balance_snapshot",
				"type",
				"purpose",
			})
			txRows.AddRow(1, 100, 1, nil, nil, "pending", "-9", nil, true, "900", "account", "fee_exchange_margin")
			txRows.AddRow(2, 100, 1, nil, nil, "pending", "-90", nil, true, "900", "account", "owt_outgoing")
			txRows.AddRow(3, 100, nil, nil, 3, "pending", "9", nil, true, "100", "account", "revenue_exchange_margin")

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

			// select revenue account
			mock.ExpectQuery("SELECT (.+) FROM `revenue_accounts`.*").
				WillReturnRows(revenueAccRows)

			type testData struct {
				id                       interface{}
				amount                   interface{}
				showAmount               interface{}
				availableBalanceSnapshot interface{}
				currentBalanceSnapshot   interface{}
			}

			testInput := []*testData{
				{
					id:                       1,
					amount:                   "-10",
					showAmount:               nil,
					availableBalanceSnapshot: "990",
					currentBalanceSnapshot:   "1000",
				},
				{
					id:                       2,
					amount:                   "-100",
					showAmount:               "-110",
					availableBalanceSnapshot: "890",
					currentBalanceSnapshot:   "1000",
				},
				{
					id:                       3,
					amount:                   "10",
					showAmount:               nil,
					availableBalanceSnapshot: "0",
					currentBalanceSnapshot:   "0",
				},
			}

			for _, data := range testInput {
				func(data *testData) {
					mock.ExpectExec("UPDATE `transactions`.*").
						//amount, available_balance_snapshot, current_balance_snapshot, show_amount, status, updated_at, id
						WithArgs(data.amount, data.availableBalanceSnapshot, data.currentBalanceSnapshot, data.showAmount, "pending", AnyTime{}, data.id).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}(data)
			}

			// update request
			mock.ExpectExec("UPDATE `requests`.*").
				WithArgs("100", "1", 100).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// update source account
			mock.ExpectExec("UPDATE `accounts`.*").
				// available_amount, balance, id
				WithArgs(str2Dec("890"), str2Dec("1000"), 1).
				WillReturnResult(sqlmock.NewResult(1, 1))

			rqs := owtRequest("100", "EUR", "USD")
			rqs.Status = pointer.ToString("pending")
			rqs.Rate = pointer.ToDecimal(str2Dec("1"))
			rqs.GetInput().Set("sourceAccountId", 1)
			rqs.GetInput().Set("revenueAccountId", 3)
			rqs.GetInput().Set("exchangeMarginPercent", "10")
			rqs.GetInput().Set("transferFeeParams", nil)
			rqs.GetInput().Set("refMessage", "")
			rqs.GetInput().Set("beneficiaryCustomerAccountName", "")

			input := NewDbOWTInput(tx, rqs, nil)
			owt := NewOutgoingWireTransfer(input, currencyBox, tx, mockPF)

			_, err := owt.Modify(rqs)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ensureTransactionsOrder(owt.Transactions())).To(Succeed())
		})

		It("should verify db input cache (should not use db if cache is provided)", func() {
			feeParams := &fee.TransferFeeParams{
				Base:    str2Dec("1"),
				Percent: str2Dec("2"),
				Min:     str2Dec("3"),
				Max:     str2Dec("4"),
			}
			cache := &OWTInputCache{
				SourceAccount:                  sourceAccountEur,
				RevenueAccount:                 revenueAccountEur,
				ExchangeMarginPercent:          pointer.ToDecimal(str2Dec("15")),
				TransferFeeParams:              feeParams,
				BeneficiaryCustomerAccountName: pointer.ToString("BeneficiaryCustomerAccountName"),
				RefMessage:                     pointer.ToString("RefMessage"),
			}
			input := NewDbOWTInput(nil, nil, cache)

			Expect(input.SourceAccount()).To(Equal(sourceAccountEur))
			Expect(input.RevenueAccount()).To(Equal(revenueAccountEur))
			Expect(input.BeneficiaryCustomerAccountName()).To(Equal("BeneficiaryCustomerAccountName"))
			Expect(input.RefMessage()).To(Equal("RefMessage"))
			Expect(input.ExchangeMarginPercent()).To(decEqual(str2Dec("15")))
			Expect(input.TransferFeeParams()).To(Equal(feeParams))
		})

	})
})
