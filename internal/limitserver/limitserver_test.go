package limitserver_test

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	mockLimit "github.com/Confialink/wallet-accounts/internal/limit/mock"
	. "github.com/Confialink/wallet-accounts/internal/limitserver"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	rpcLimit "github.com/Confialink/wallet-accounts/rpc/limit"
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var _ = Describe("Limit PRC server", func() {
	Context("server", func() {
		var (
			mock sqlmock.Sqlmock
			gdb  *gorm.DB
		)

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

		When("invalid amount value is given", func() {
			It("must return error", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				limStorage := mockLimit.NewMockStorage(ctrl)
				limitService := limit.NewService(limStorage, limit.NewFactory())

				server := NewServer(limitService, gdb)
				request := &rpcLimit.SetLimitsRequest{
					Limits: []*rpcLimit.LimitWithId{
						{
							Limit: &rpcLimit.Limit{
								Amount: "unexpected value",
							},
						},
					},
				}

				mock.ExpectBegin()
				mock.ExpectRollback()

				_, err := server.Set(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to convert"))
			})
		})

		When("limit identifier is incomplete", func() {
			Specify("limits cannot be set", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				limStorage := mockLimit.NewMockStorage(ctrl)
				limitService := limit.NewService(limStorage, limit.NewFactory())
				server := NewServer(limitService, gdb)

				request := &rpcLimit.SetLimitsRequest{
					Limits: []*rpcLimit.LimitWithId{
						{
							Limit: &rpcLimit.Limit{
								CurrencyCode: "EUR",
								Amount:       "199.99",
							},
							LimitId: &rpcLimit.LimitId{
								Name:     rpcLimit.LimitName_EMPTY, // not filled i.e. id is incomplete
								Entity:   "user",
								EntityId: "123",
							},
						},
					},
				}
				mock.ExpectBegin()
				mock.ExpectRollback()

				_, err := server.Set(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(limit.ErrIdIncomplete))
			})
			Specify("limit cannot be reset", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				limStorage := mockLimit.NewMockStorage(ctrl)
				limitService := limit.NewService(limStorage, limit.NewFactory())
				server := NewServer(limitService, gdb)

				request := &rpcLimit.ResetLimitsRequest{
					Identifiers: []*rpcLimit.LimitId{
						{
							Name:     rpcLimit.LimitName_MAX_TOTAL_DEBIT_PER_MONTH,
							Entity:   "", // not filled i.e. id is incomplete
							EntityId: "123",
						},
					},
				}
				mock.ExpectBegin()
				mock.ExpectRollback()

				_, err := server.ResetToDefault(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(limit.ErrIdIncomplete))
			})
		})

		When("limit is already exist", func() {
			It("could be updated", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				id := limit.Identifier{
					Name:     transfers.LimitMaxCreditPerTransfer,
					Entity:   "user",
					EntityId: "mock_user_id",
				}

				limStorage := mockLimit.NewMockStorage(ctrl)
				limStorage.
					EXPECT().
					Find(id).
					Return([]limit.Model{
						// returning value does not matter
						// the main thing is the fact that the value is returned (means that it is exist)
						{Value: limit.NoLimit()},
					}, nil)
				limStorage.
					EXPECT().
					Update(limit.Val(dec(199.99), "EUR"), id).
					Return(nil)

				limitService := limit.NewService(limStorage, limit.NewFactory())
				server := NewServer(limitService, gdb)

				request := &rpcLimit.SetLimitsRequest{
					Limits: []*rpcLimit.LimitWithId{
						{
							Limit: &rpcLimit.Limit{
								CurrencyCode: "EUR",
								Amount:       "199.99",
							},
							LimitId: &rpcLimit.LimitId{
								Name:     rpcLimit.LimitName_MAX_CREDIT_PER_TRANSFER,
								Entity:   "user",
								EntityId: "mock_user_id",
							},
						},
					},
				}
				mock.ExpectBegin()
				mock.ExpectCommit()

				_, err := server.Set(context.Background(), request)
				Expect(err).ToNot(HaveOccurred())
			})
			It("could be found", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				// id is not complete which means that multiple limits could be found
				id := limit.Identifier{
					Name:   transfers.LimitMaxDebitPerTransfer,
					Entity: "user",
				}
				val1 := limit.Val(dec(12345.67), "BTC")
				val2 := limit.NoLimit()
				limStorage := mockLimit.NewMockStorage(ctrl)
				limStorage.
					EXPECT().
					Find(id).
					Return([]limit.Model{
						{
							Identifier: limit.Identifier{
								Name:     id.Name,
								Entity:   id.Entity,
								EntityId: "mock_user_1",
							},
							Value: val1,
						},
						{
							Identifier: limit.Identifier{
								Name:     id.Name,
								Entity:   id.Entity,
								EntityId: "mock_user_2",
							},
							Value: val2,
						},
					}, nil). // record is not found
					AnyTimes()

				limitService := limit.NewService(limStorage, limit.NewFactory())
				server := NewServer(limitService, gdb)

				request := &rpcLimit.GetLimitsRequest{
					// id is not complete which means that multiple limits could be found
					Identifiers: []*rpcLimit.LimitId{
						{
							Name:   rpcLimit.LimitName_MAX_DEBIT_PER_TRANSFER,
							Entity: "user",
						},
					},
				}

				response, err := server.Get(context.Background(), request)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Limits).To(HaveLen(2))

				lim1, lim2 := response.Limits[0], response.Limits[1]
				reqId := request.Identifiers[0]

				Expect(lim1.Limit.Exists).To(BeTrue())
				Expect(lim1.LimitId).To(Equal(&rpcLimit.LimitId{
					Name:     reqId.Name,
					Entity:   reqId.Entity,
					EntityId: "mock_user_1",
				}))
				Expect(lim1.Limit.NoLimit).To(BeFalse())
				Expect(lim1.Limit.CurrencyCode).To(Equal(val1.CurrencyAmount().CurrencyCode()))
				Expect(lim1.Limit.Amount).To(Equal(val1.CurrencyAmount().Amount().String()))

				Expect(lim2.Limit.Exists).To(BeTrue())
				Expect(lim2.LimitId).To(Equal(&rpcLimit.LimitId{
					Name:     reqId.Name,
					Entity:   reqId.Entity,
					EntityId: "mock_user_2",
				}))
				Expect(lim2.Limit.NoLimit).To(BeTrue())
				Expect(lim2.Limit.CurrencyCode).To(Equal(""))
				Expect(lim2.Limit.Amount).To(Equal(""))

			})
		})

		When("limit is not exist", func() {
			It("could be created", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				id := limit.Identifier{
					Name:     transfers.LimitMaxTotalBalance,
					Entity:   "user",
					EntityId: "mock_user_id",
				}

				limStorage := mockLimit.NewMockStorage(ctrl)
				limStorage.
					EXPECT().
					Find(id).
					Return(nil, limit.ErrNotFound). // record is not found
					AnyTimes()

				limStorage.
					EXPECT().
					Save(limit.Val(dec(199.99), "EUR"), id).
					Return(nil)

				limitService := limit.NewService(limStorage, limit.NewFactory())
				server := NewServer(limitService, gdb)

				request := &rpcLimit.SetLimitsRequest{
					Limits: []*rpcLimit.LimitWithId{
						{
							Limit: &rpcLimit.Limit{
								CurrencyCode: "EUR",
								Amount:       "199.99",
							},
							LimitId: &rpcLimit.LimitId{
								Name:     rpcLimit.LimitName_MAX_TOTAL_BALANCE,
								Entity:   "user",
								EntityId: "mock_user_id",
							},
						},
					},
				}
				mock.ExpectBegin()
				mock.ExpectCommit()

				_, err := server.Set(context.Background(), request)
				Expect(err).ToNot(HaveOccurred())
			})
			It("should be marked", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				id := limit.Identifier{
					Name:     transfers.LimitMaxDebitPerTransfer,
					Entity:   "user",
					EntityId: "mock_user_id",
				}

				limStorage := mockLimit.NewMockStorage(ctrl)
				limStorage.
					EXPECT().
					Find(id).
					Return(nil, limit.ErrNotFound). // record is not found
					AnyTimes()

				limitService := limit.NewService(limStorage, limit.NewFactory())
				server := NewServer(limitService, gdb)

				request := &rpcLimit.GetLimitsRequest{
					Identifiers: []*rpcLimit.LimitId{
						{
							Name:     rpcLimit.LimitName_MAX_DEBIT_PER_TRANSFER,
							Entity:   "user",
							EntityId: "mock_user_id",
						},
					},
				}

				response, err := server.Get(context.Background(), request)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Limits).To(HaveLen(1))

				lim := response.Limits[0]
				Expect(lim.Limit.Exists).To(BeFalse())
				Expect(lim.LimitId).To(Equal(request.Identifiers[0]))
			})
		})

		Specify("limit could be reset to default using 'complete id'", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			limStorage := mockLimit.NewMockStorage(ctrl)
			limStorage.
				EXPECT().
				Delete(limit.Identifier{
					Name:     transfers.LimitMaxTotalDebitPerMonth,
					Entity:   "user",
					EntityId: "123",
				})

			limitService := limit.NewService(limStorage, limit.NewFactory())
			server := NewServer(limitService, gdb)

			request := &rpcLimit.ResetLimitsRequest{
				Identifiers: []*rpcLimit.LimitId{
					{
						Name:     rpcLimit.LimitName_MAX_TOTAL_DEBIT_PER_MONTH,
						Entity:   "user",
						EntityId: "123",
					},
				},
			}
			mock.ExpectBegin()
			mock.ExpectCommit()

			_, err := server.ResetToDefault(context.Background(), request)
			Expect(err).ToNot(HaveOccurred())
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
