package transfers_test

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	"github.com/Confialink/wallet-accounts/internal/limit"
	mockLimit "github.com/Confialink/wallet-accounts/internal/limit/mock"
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	mockBalance "github.com/Confialink/wallet-accounts/internal/modules/balance/mock"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	. "github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	mockTransfers "github.com/Confialink/wallet-accounts/internal/modules/request/transfers/mock"
	txConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/inconshreveable/log15"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"time"
)

type testPermission struct {
	shouldFail bool
	msg        string
}

func (s testPermission) Name() string {
	return "test"
}

func (s testPermission) Check() error {
	if s.shouldFail {
		return errors.New(s.msg)
	}
	return nil
}

var _ = Describe("Transfers", func() {
	Context("Common transfer logic", func() {
		When("withdrawal is not allowed", func() {
			It("should raise error", func() {
				sourceAcc := account("EUR", "100")
				sourceAcc.AllowWithdrawals = pointer.ToBool(true)

				withdrawalAllowedRule := NewWithdrawalPermission(sourceAcc)
				Expect(withdrawalAllowedRule.Check()).To(Succeed())

				sourceAcc.AllowWithdrawals = pointer.ToBool(false)
				err := withdrawalAllowedRule.Check()
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrWithdrawalNotAllowed))

				sourceAcc.AllowWithdrawals = nil
				err = withdrawalAllowedRule.Check()
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrWithdrawalNotAllowed))
			})
		})
		When("deposit is not allowed", func() {
			It("should raise error", func() {
				destAcc := account("EUR", "100")
				destAcc.AllowDeposits = pointer.ToBool(true)

				depositAllowedRule := NewDepositPermission(destAcc)
				Expect(depositAllowedRule.Check()).To(Succeed())

				destAcc.AllowDeposits = pointer.ToBool(false)
				err := depositAllowedRule.Check()
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrDepositNotAllowed))

				destAcc.AllowDeposits = nil
				err = depositAllowedRule.Check()
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrDepositNotAllowed))
			})
		})
		When("account is not active", func() {
			It("should raise error", func() {
				acc := account("EUR", "100")
				accountActivePermission := NewAccountActivePermission(acc)

				acc.IsActive = pointer.ToBool(true)
				Expect(accountActivePermission.Check()).To(Succeed())

				acc.IsActive = pointer.ToBool(false)
				err := accountActivePermission.Check()
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrAccountInactive))

				acc.AllowDeposits = nil
				err = accountActivePermission.Check()
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrAccountInactive))
			})
		})
		It(`should check "SufficientBalancePermission"`, func() {
			type testData struct {
				requested int64
				available int64
				err       bool
			}
			data := []testData{
				{100, 100, false},
				{100, 500, false},
				{500, 400, true},
				{1, -1, true},
				{-1, 2, false},
			}
			for _, v := range data {
				func(v testData) {
					requested := SimpleAmountable(decimal.NewFromInt(v.requested))
					available := SimpleAmountable(decimal.NewFromInt(v.available))

					rule := NewSufficientBalancePermission(requested, available)
					err := rule.Check()
					if v.err {
						Expect(err).To(HaveOccurred())
						Expect(errors.Cause(err)).To(Equal(ErrInsufficientBalance))
						return
					}
					Expect(err).ShouldNot(HaveOccurred())
				}(v)
			}
		})

		It("combines multiple permissions", func() {
			permissions := PermissionCheckers{
				&testPermission{
					shouldFail: false,
				},
				&testPermission{
					shouldFail: true,
					msg:        "first",
				},
				&testPermission{
					shouldFail: true,
					msg:        "second",
				},
			}

			err := permissions.Check()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("first"))

			permissions = PermissionCheckers{
				&testPermission{
					shouldFail: false,
				},
				&testPermission{
					shouldFail: false,
				},
				&testPermission{
					shouldFail: true,
					msg:        "second",
				},
			}

			err = permissions.Check()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("second"))

			permissions = PermissionCheckers{
				&testPermission{
					shouldFail: false,
				},
			}
			Expect(permissions.Check()).To(Succeed())
		})

		It("should ensure rate base and reference codes are based on rate designation", func() {
			req := request("1", "EUR", "USD")
			Expect(req.RateBaseCurrencyCode()).To(Equal("EUR"))
			Expect(req.RateReferenceCurrencyCode()).To(Equal("USD"))

			req.RateDesignation = model.RateDesignationReferenceBase
			Expect(req.RateBaseCurrencyCode()).To(Equal("USD"))
			Expect(req.RateReferenceCurrencyCode()).To(Equal("EUR"))

			req.RateDesignation = model.RateDesignationBaseReference
			Expect(req.RateBaseCurrencyCode()).To(Equal("EUR"))
			Expect(req.RateReferenceCurrencyCode()).To(Equal("USD"))

			req.RateDesignation = "anything"
			Expect(req.RateBaseCurrencyCode()).To(Equal("EUR"))
			Expect(req.RateReferenceCurrencyCode()).To(Equal("USD"))
		})

		It("should create executor/canceller/modifier by request subject", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			mockPF := mockTransfers.NewMockPermissionFactory(ctrl)
			mockPF.
				EXPECT().
				WrapContext(gomock.Any()).
				AnyTimes()

			rqs := request("1", "EUR")

			subjects := []constants.Subject{
				"TBA", "TBU", "OWT",
			}
			for _, subject := range subjects {
				func(subject constants.Subject) {
					note := fmt.Sprintf("subject %s", subject)
					rqs.Subject = &subject
					_, err := CreateExecutor(nil, rqs, currencyBox, mockPF)
					Expect(err).ToNot(HaveOccurred(), "executor "+note)

					_, err = CreateCanceller(nil, rqs, currencyBox, mockPF)
					Expect(err).ToNot(HaveOccurred(), "canceller "+note)

					_, err = CreateModifier(nil, rqs, currencyBox, mockPF)
					Expect(err).ToNot(HaveOccurred(), "modifier "+note)
				}(subject)
			}
		})

		It("should ensure that transfer fee params could be retrieved from request", func() {
			rqs := request("1", "CHF")
			params := &fee.TransferFeeParams{
				Base:    str2Dec("1"),
				Percent: str2Dec("2"),
				Min:     str2Dec("3"),
				Max:     str2Dec("4"),
			}

			type testData struct {
				inputValue    interface{}
				errorExpected bool
				matcher       OmegaMatcher
			}

			tests := []testData{
				{
					inputValue:    params,
					errorExpected: false,
					matcher:       Equal(params),
				},
				{
					inputValue:    *params,
					errorExpected: false,
					matcher:       Equal(params),
				},
				{
					inputValue:    nil,
					errorExpected: false,
					matcher:       BeNil(),
				},
				{
					inputValue: map[string]interface{}{
						"base":    "1",
						"percent": str2Dec("2"),
						"min":     pointer.ToDecimal(str2Dec("3")),
						"max":     "4",
					},
					errorExpected: false,
					matcher:       Equal(params),
				},
				{
					inputValue: map[string]interface{}{
						"percent": str2Dec("2"),
						"min":     pointer.ToDecimal(str2Dec("3")),
						"max":     "4",
					},
					errorExpected: true,
					matcher:       BeNil(),
				},
				{
					inputValue: map[string]interface{}{
						"base": "1",
						"min":  pointer.ToDecimal(str2Dec("3")),
						"max":  "4",
					},
					errorExpected: true,
					matcher:       BeNil(),
				},
				{
					inputValue: map[string]interface{}{
						"base":    "1",
						"percent": str2Dec("2"),
						"max":     "4",
					},
					errorExpected: true,
					matcher:       BeNil(),
				},
				{
					inputValue: map[string]interface{}{
						"base":    "1",
						"percent": str2Dec("2"),
						"min":     pointer.ToDecimal(str2Dec("3")),
					},
					errorExpected: true,
					matcher:       BeNil(),
				},
			}

			errMatcher := map[bool]OmegaMatcher{
				true:  HaveOccurred(),
				false: Not(HaveOccurred()),
			}
			for _, data := range tests {
				func(t testData) {
					rqs.GetInput().Set("transferFeeParams", t.inputValue)
					baInput := NewDbBetweenAccountsInput(nil, rqs, nil)
					owtInput := NewDbOWTInput(nil, rqs, nil)

					params, err := baInput.TransferFeeParams()
					Expect(err).To(errMatcher[t.errorExpected])
					Expect(params).To(t.matcher)
					params, err = owtInput.TransferFeeParams()
					Expect(err).To(errMatcher[t.errorExpected])
					Expect(params).To(t.matcher)
				}(data)
			}
		})
	})

	Context("Transfer limitations", func() {
		It("should check limit storage decorator", func() {
			// these errors are used as samples to ensure that the decorator passes calls to the appropriate methods
			var (
				errSave   = errors.New("save")
				errUpdate = errors.New("update")
				errDelete = errors.New("delete")
			)
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			any := gomock.Any()
			mockStorage := mockLimit.NewMockStorage(ctrl)
			mockStorage.
				EXPECT().
				Save(any, any).
				Return(errSave)
			mockStorage.
				EXPECT().
				Update(any, any).
				Return(errUpdate)
			mockStorage.
				EXPECT().
				Delete(any).
				Return(errDelete)

			mockStorage.
				EXPECT().
				Find(any).
				Return(nil, limit.ErrNotFound).
				AnyTimes().
				After(
					mockStorage.
						EXPECT().
						Find(any).
						Return(make([]limit.Model, 0), nil).
						MaxTimes(4),
				)
			decoratorStorage := NewLimitStorageDecorator(mockStorage)

			val, id := limit.Val(dec(1), "ANY"), limit.Identifier{}
			err := decoratorStorage.Save(val, id)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errSave))

			err = decoratorStorage.Update(val, id)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errUpdate))

			err = decoratorStorage.Delete(id)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errDelete))

			names := []string{
				LimitMaxTotalBalance,
				LimitMaxDebitPerTransfer,
				LimitMaxTotalDebitPerDay,
				LimitMaxTotalDebitPerMonth,
				// repeated b/c mock storage should return different result: ErrNotFound and then nil error
				LimitMaxTotalBalance,
				LimitMaxDebitPerTransfer,
				LimitMaxTotalDebitPerDay,
				LimitMaxTotalDebitPerMonth,
			}

			type defaultVal struct {
				currency string
				val      float64
			}
			defaultValues := map[string]defaultVal{
				LimitMaxTotalBalance: {
					LimitMaxTotalBalanceDefaultCurrency,
					LimitMaxTotalBalanceDefaultAmount,
				},
				LimitMaxDebitPerTransfer: {
					LimitMaxDebitPerTransferDefaultCurrency,
					LimitMaxDebitPerTransferDefaultAmount,
				},
				LimitMaxTotalDebitPerDay: {
					LimitMaxTotalDebitPerDayDefaultCurrency,
					LimitMaxTotalDebitPerDayDefaultAmount,
				},
				LimitMaxTotalDebitPerMonth: {
					LimitMaxTotalDebitPerMonthDefaultCurrency,
					LimitMaxTotalDebitPerMonthDefaultAmount,
				},
			}

			for _, name := range names {
				func(name string) {
					id := limit.Identifier{Name: name, Entity: "any", EntityId: "any"}
					models, err := decoratorStorage.Find(id)
					Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("name: %s", name))
					Expect(models).To(HaveLen(1))
					lim := models[0]
					Expect(lim.Name).To(Equal(name))
					defVal := defaultValues[name]
					Expect(lim.Value.CurrencyAmount().Amount()).To(decEqual(dec(defVal.val)))
					Expect(lim.Value.CurrencyAmount().CurrencyCode()).To(Equal(defVal.currency))
				}(name)
			}
		})

		It("should check max balance limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			acc := account("USD", "1000")
			acc.UserId = "user1"
			// as if we credit 300 + 500 USD
			details := types.Details{
				txConstants.Purpose("credit1"): {
					Amount:       dec(300),
					CurrencyCode: "USD",
					Account:      acc,
				},
				txConstants.Purpose("credit2"): {
					Amount:       dec(500),
					CurrencyCode: "USD",
					Account:      acc,
				},
			}

			limitStorage := mockLimit.NewMockStorage(ctrl)
			limitStorage.
				EXPECT().
				Find(limit.Identifier{
					Name:     LimitMaxTotalBalance,
					Entity:   "user",
					EntityId: "user1",
				}).Return([]limit.Model{
				{
					Identifier: limit.Identifier{
						Name:     LimitMaxTotalBalance,
						Entity:   "user",
						EntityId: "user1",
					},
					Value: limit.Val(dec(1000), "EUR"), // allowed maximum is 1000EUR
				},
			}, nil).
				AnyTimes()

			limitService := limit.NewService(limitStorage, limit.NewFactory())
			rateSource := exchange.NewDirectRateSource()
			// for simplicity rate is 1/1
			_ = rateSource.Set(exchange.NewRate("EUR", "USD", dec(1)))
			_ = rateSource.Set(exchange.NewRate("USD", "EUR", dec(1)))

			reducer := balance.NewDefaultReducer(rateSource)
			// total user1 balances are 100EUR + 100USD
			totalBalanceResult := balance.AggregationResult{
				{ItemAmount: dec(100), ItemCurrencyCode: "EUR"},
				{ItemAmount: dec(100), ItemCurrencyCode: "USD"},
			}
			totalBalanceAggregator := mockBalance.NewMockAggregator(ctrl)
			totalBalanceAggregator.
				EXPECT().
				Aggregate().
				Return(totalBalanceResult, nil).
				AnyTimes()

			aggregationFactory := mockBalance.NewMockAggregationFactory(ctrl)
			aggregationFactory.
				EXPECT().
				GeneralTotalByUserId("user1").
				Return(totalBalanceAggregator, nil).
				AnyTimes()

			aggregationService := balance.NewAggregationService(reducer, aggregationFactory)

			maxBalancePermission := NewMaxBalanceLimit(details, limitService, aggregationService, &mockLogger{})
			// total balance + credit details should give exactly 1000EUR which is allowed maximum
			Expect(maxBalancePermission.Check()).To(Succeed())

			// now lets try to add more than allowed
			details[txConstants.Purpose("credit3")] = &types.Detail{
				Amount:       dec(1),
				CurrencyCode: "USD",
				Account:      acc,
			}

			maxBalancePermission = NewMaxBalanceLimit(details, limitService, aggregationService, &mockLogger{})
			err := maxBalancePermission.Check()
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(limit.ErrLimitExceeded))

			noLimit := mockLimit.NewMockValue(ctrl)
			noLimit.
				EXPECT().
				NoLimit().
				Return(true).
				AnyTimes()

			noLimitStorage := mockLimit.NewMockStorage(ctrl)
			noLimitStorage.
				EXPECT().
				Find(gomock.Any()).
				Return(
					[]limit.Model{
						{
							Identifier: limit.Identifier{},
							Value:      noLimit,
						},
					},
					nil,
				).
				AnyTimes()
			noLimitService := limit.NewService(noLimitStorage, limit.NewFactory())

			maxBalancePermission = NewMaxBalanceLimit(details, noLimitService, aggregationService, &mockLogger{})
			Expect(maxBalancePermission.Check()).To(Succeed())

		})

		It("should check max total debit per transfer limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			acc := account("USD", "10000")
			acc.UserId = "user1"
			// as if we debit (-300) + (-500) USD
			details := types.Details{
				txConstants.Purpose("debit1"): {
					Amount:       dec(-300),
					CurrencyCode: "USD",
					Account:      acc,
				},
				txConstants.Purpose("debit2"): {
					Amount:       dec(-500),
					CurrencyCode: "USD",
					Account:      acc,
				},
			}

			limitStorage := mockLimit.NewMockStorage(ctrl)
			limitStorage.
				EXPECT().
				Find(limit.Identifier{
					Name:     LimitMaxDebitPerTransfer,
					Entity:   "user",
					EntityId: "user1",
				}).Return([]limit.Model{
				{
					Identifier: limit.Identifier{
						Name:     LimitMaxDebitPerTransfer,
						Entity:   "user",
						EntityId: "user1",
					},
					Value: limit.Val(dec(1000), "EUR"), // allowed maximum is 1000EUR
				},
			}, nil).
				AnyTimes()

			limitService := limit.NewService(limitStorage, limit.NewFactory())

			rateSource := exchange.NewDirectRateSource()
			// for simplicity rate is 1/1
			_ = rateSource.Set(exchange.NewRate("EUR", "USD", dec(1)))
			_ = rateSource.Set(exchange.NewRate("USD", "EUR", dec(1)))

			reducer := balance.NewDefaultReducer(rateSource)

			maxDebitPermission := NewMaxDebitPerTransfer(details, limitService, reducer, &mockLogger{})
			Expect(maxDebitPermission.Check()).To(Succeed())

			// now lets try to add more than allowed
			details[txConstants.Purpose("debit3")] = &types.Detail{
				Amount:       dec(-201),
				CurrencyCode: "USD",
				Account:      acc,
			}

			maxDebitPermission = NewMaxDebitPerTransfer(details, limitService, reducer, &mockLogger{})
			err := maxDebitPermission.Check()
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(limit.ErrLimitExceeded))

			noLimit := mockLimit.NewMockValue(ctrl)
			noLimit.
				EXPECT().
				NoLimit().
				Return(true).
				AnyTimes()

			noLimitStorage := mockLimit.NewMockStorage(ctrl)
			noLimitStorage.
				EXPECT().
				Find(gomock.Any()).
				Return(
					[]limit.Model{
						{
							Identifier: limit.Identifier{},
							Value:      noLimit,
						},
					},
					nil,
				).
				AnyTimes()
			noLimitService := limit.NewService(noLimitStorage, limit.NewFactory())

			maxDebitPermission = NewMaxDebitPerTransfer(details, noLimitService, reducer, &mockLogger{})
			Expect(maxDebitPermission.Check()).To(Succeed())
		})

		It("should check max total credit per transfer limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			acc := account("USD", "10000")
			acc.UserId = "user1"

			details := types.Details{
				txConstants.Purpose("credit1"): {
					Amount:       dec(300),
					CurrencyCode: "USD",
					Account:      acc,
				},
				txConstants.Purpose("credit2"): {
					Amount:       dec(500),
					CurrencyCode: "USD",
					Account:      acc,
				},
			}

			limitStorage := mockLimit.NewMockStorage(ctrl)
			limitStorage.
				EXPECT().
				Find(limit.Identifier{
					Name:     LimitMaxCreditPerTransfer,
					Entity:   "user",
					EntityId: "user1",
				}).Return([]limit.Model{
				{
					Identifier: limit.Identifier{
						Name:     LimitMaxCreditPerTransfer,
						Entity:   "user",
						EntityId: "user1",
					},
					Value: limit.Val(dec(1000), "EUR"), // allowed maximum is 1000EUR
				},
			}, nil).
				AnyTimes()

			limitService := limit.NewService(limitStorage, limit.NewFactory())

			rateSource := exchange.NewDirectRateSource()
			// for simplicity rate is 1/1
			_ = rateSource.Set(exchange.NewRate("EUR", "USD", dec(1)))
			_ = rateSource.Set(exchange.NewRate("USD", "EUR", dec(1)))

			reducer := balance.NewDefaultReducer(rateSource)

			maxCreditPermission := NewMaxCreditPerTransfer(details, limitService, reducer, &mockLogger{})
			Expect(maxCreditPermission.Check()).To(Succeed())

			// now lets try to add more than allowed
			details[txConstants.Purpose("credit3")] = &types.Detail{
				Amount:       dec(201),
				CurrencyCode: "USD",
				Account:      acc,
			}

			maxCreditPermission = NewMaxCreditPerTransfer(details, limitService, reducer, &mockLogger{})
			err := maxCreditPermission.Check()
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(limit.ErrLimitExceeded))

			noLimit := mockLimit.NewMockValue(ctrl)
			noLimit.
				EXPECT().
				NoLimit().
				Return(true).
				AnyTimes()

			noLimitStorage := mockLimit.NewMockStorage(ctrl)
			noLimitStorage.
				EXPECT().
				Find(gomock.Any()).
				Return(
					[]limit.Model{
						{
							Identifier: limit.Identifier{},
							Value:      noLimit,
						},
					},
					nil,
				).
				AnyTimes()
			noLimitService := limit.NewService(noLimitStorage, limit.NewFactory())
			maxCreditPermission = NewMaxCreditPerTransfer(details, noLimitService, reducer, &mockLogger{})

			Expect(maxCreditPermission.Check()).To(Succeed())
		})

		It("should check max total debit per period limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			acc := account("USD", "1000")
			acc.UserId = "user1"
			// as if we debit (-300) + (-500 USD)
			details := types.Details{
				txConstants.Purpose("debit1"): {
					Amount:       dec(-300),
					CurrencyCode: "USD",
					Account:      acc,
				},
				txConstants.Purpose("debit2"): {
					Amount:       dec(-500),
					CurrencyCode: "USD",
					Account:      acc,
				},
			}

			limitName := "max_total_debit_per_generic_period"
			limitStorage := mockLimit.NewMockStorage(ctrl)
			limitStorage.
				EXPECT().
				Find(limit.Identifier{
					Name:     limitName,
					Entity:   "user",
					EntityId: "user1",
				}).Return([]limit.Model{
				{
					Identifier: limit.Identifier{
						Name:     limitName,
						Entity:   "user",
						EntityId: "user1",
					},
					Value: limit.Val(dec(1000), "EUR"), // allowed maximum is 1000EUR
				},
			}, nil).
				AnyTimes()
			limitService := limit.NewService(limitStorage, limit.NewFactory())

			rateSource := exchange.NewDirectRateSource()
			// for simplicity rate is 1/1
			_ = rateSource.Set(exchange.NewRate("EUR", "USD", dec(1)))
			_ = rateSource.Set(exchange.NewRate("USD", "EUR", dec(1)))

			reducer := balance.NewDefaultReducer(rateSource)

			totalPerPeriodResult := balance.AggregationResult{
				{ItemAmount: dec(100), ItemCurrencyCode: "EUR"},
				{ItemAmount: dec(100), ItemCurrencyCode: "USD"},
			}
			totalPerPeriodAggregator := mockBalance.NewMockAggregator(ctrl)
			totalPerPeriodAggregator.
				EXPECT().
				Aggregate().
				Return(totalPerPeriodResult, nil).
				AnyTimes()

			aggregationFactory := mockBalance.NewMockAggregationFactory(ctrl)
			aggregationFactory.
				EXPECT().
				TotalDebitedByUserIdPerPeriod("user1", gomock.Any(), gomock.Any()).
				Return(totalPerPeriodAggregator, nil).
				MaxTimes(2)

			aggregationService := balance.NewAggregationService(reducer, aggregationFactory)

			totalDebitPerPeriodPermission := NewMaxTotalDebitPerPeriod(
				details,
				limitService,
				aggregationService,
				limitName,
				time.Time{},
				time.Time{},
				&mockLogger{},
			)

			Expect(totalDebitPerPeriodPermission.Check()).To(Succeed())

			// now lets try to add more than allowed
			details[txConstants.Purpose("debit3")] = &types.Detail{
				Amount:       dec(-1),
				CurrencyCode: "USD",
				Account:      acc,
			}

			totalDebitPerPeriodPermission = NewMaxTotalDebitPerPeriod(
				details,
				limitService,
				aggregationService,
				limitName,
				time.Time{},
				time.Time{},
				&mockLogger{},
			)
			err := totalDebitPerPeriodPermission.Check()
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(limit.ErrLimitExceeded))

			noLimit := mockLimit.NewMockValue(ctrl)
			noLimit.
				EXPECT().
				NoLimit().
				Return(true).
				AnyTimes()

			noLimitStorage := mockLimit.NewMockStorage(ctrl)
			noLimitStorage.
				EXPECT().
				Find(gomock.Any()).
				Return(
					[]limit.Model{
						{
							Identifier: limit.Identifier{},
							Value:      noLimit,
						},
					},
					nil,
				).
				AnyTimes()
			noLimitService := limit.NewService(noLimitStorage, limit.NewFactory())
			totalDebitPerPeriodPermission = NewMaxTotalDebitPerPeriod(
				details,
				noLimitService,
				aggregationService,
				limitName,
				time.Time{},
				time.Time{},
				&mockLogger{},
			)

			Expect(totalDebitPerPeriodPermission.Check()).To(Succeed())
		})

		It("should verify default permissions", func() {
			acc1 := account("USD", "1000")
			acc1.UserId = "user1"

			acc2 := account("USD", "1000")
			acc2.UserId = "user2"

			defaultPf := NewDefaultPermissionFactory(nil, nil, nil, nil, nil)

			details := types.Details{
				txConstants.Purpose("debit1"): {
					Amount:       dec(-300),
					CurrencyCode: "USD",
					Account:      acc1,
				},
				txConstants.Purpose("credit1"): {
					Amount:       dec(300),
					CurrencyCode: "USD",
					Account:      acc2,
				},
			}
			req := request("300", "USD")

			permissions, err := defaultPf.CreatePermission(req, details)
			Expect(err).ToNot(HaveOccurred())

			btoi := map[bool]int{false: 0, true: 1}
			// int is a number of expected entries of the permission
			expectedPermissions := map[string]int{
				"deposit_allowed":          1,
				"withdrawal_allowed":       1,
				"sufficient_balance":       1,
				"account_active":           2,
				LimitMaxTotalBalance:       btoi[LimitMaxTotalBalanceEnabled],
				LimitMaxDebitPerTransfer:   btoi[LimitMaxDebitPerTransferEnabled],
				LimitMaxCreditPerTransfer:  btoi[LimitMaxCreditPerTransferEnabled],
				LimitMaxTotalDebitPerMonth: btoi[LimitMaxTotalDebitPerMonthEnabled],
				LimitMaxTotalDebitPerDay:   btoi[LimitMaxTotalDebitPerDayEnabled],
			}

			totalExpected := 0
			for _, c := range expectedPermissions {
				totalExpected += c
			}

			totalReceived := len(permissions.(PermissionCheckers))
			Expect(totalExpected).To(Equal(totalReceived), "expected permissions count must match received permissions count")

			for _, permission := range permissions.(PermissionCheckers) {
				func(permission PermissionChecker) {
					name := permission.Name()
					c, expected := expectedPermissions[name]
					enabled := c > 0
					Expect(expected).To(BeTrue(), fmt.Sprintf("permission %s is not expected", name))
					Expect(enabled).To(BeTrue(), fmt.Sprintf("permission %s expected to be enabled", name))
				}(permission)
			}
		})
	})
})

type mockLogger struct{}

func (m mockLogger) New(ctx ...interface{}) log15.Logger {
	return &m
}

func (m mockLogger) GetHandler() log15.Handler {
	panic("stub method: not implemented")
}

func (m mockLogger) SetHandler(h log15.Handler) {
	panic("stub method: not implemented")
}

func (m mockLogger) Debug(msg string, ctx ...interface{}) {}

func (m mockLogger) Info(msg string, ctx ...interface{}) {}

func (m mockLogger) Warn(msg string, ctx ...interface{}) {}

func (m mockLogger) Error(msg string, ctx ...interface{}) {}

func (m mockLogger) Crit(msg string, ctx ...interface{}) {}

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
