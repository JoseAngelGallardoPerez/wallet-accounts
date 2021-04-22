package limit_test

import (
	. "github.com/Confialink/wallet-accounts/internal/limit"
	mockLimit "github.com/Confialink/wallet-accounts/internal/limit/mock"
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

var _ = Describe("Limit", func() {
	var any = gomock.Any()
	Context("types of limits", func() {
		It("should represent type", func() {
			simpleLimit := New(val(1000))
			available := simpleLimit.Available()
			Expect(available.CurrencyAmount().Amount()).To(decEqual(1000))
			Expect(simpleLimit.WithinLimit(amount(900))).To(Succeed())
			Expect(simpleLimit.WithinLimit(amount(1000))).To(Succeed())

			err := simpleLimit.WithinLimit(amount(1001))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrLimitExceeded))

			err = simpleLimit.WithinLimit(amount(1, "UNK"))
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrCurrenciesMismatch))

			for _, a := range []CurrencyAmount{amount(0), amount(-1), amount("-0.1")} {
				func(a CurrencyAmount) {
					err = simpleLimit.WithinLimit(a)
					Expect(err).To(HaveOccurred())
					Expect(errors.Cause(err)).To(Equal(ErrInvalidAmount))
				}(a)
			}
		})

		When("not limited", func() {
			It("should accept any values without error", func() {
				lim := New(NoLimit())
				Expect(lim.Available().NoLimit()).To(BeTrue())
				Expect(lim.Available().CurrencyAmount()).To(BeNil())

				testData := []CurrencyAmount{
					amount(1, "EUR"),
					amount(-999999, "EUR"),
					amount(9999999999, "USD"),
					amount(.1, "ANYTHING"),
				}

				for _, t := range testData {
					func(t CurrencyAmount) {
						Expect(lim.WithinLimit(t)).To(Succeed())
					}(t)
				}
			})
		})

		It("should check identifier uniqueness", func() {
			type testData struct {
				id     Identifier
				unique bool
			}
			cases := []testData{
				{Identifier{"name", "entity", "id"}, true},
				{Identifier{"", "entity", "id"}, false},
				{Identifier{"name", "", "id"}, false},
				{Identifier{"name", "entity", ""}, false},
				{Identifier{"", "", "id"}, false},
				{Identifier{"", "", ""}, false},
				{Identifier{"name", "", ""}, false},
				{Identifier{" ", " ", " "}, true},
				{Identifier{" ", " ", "0"}, true},
			}
			for _, t := range cases {
				func(t testData) {
					id := t.id
					note := fmt.Sprintf("name: '%s', entity: '%s', id: '%s'", id.Name, id.Entity, id.EntityId)
					Expect(id.IsUnique()).To(Equal(t.unique), note)
				}(t)
			}

		})
	})
	Context("storage and service", func() {
		It("should save limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()
			storage := mockLimit.NewMockStorage(ctrl)
			srv := NewService(storage, NewFactory())

			id := Identifier{
				Name:     "limit1",
				Entity:   "user",
				EntityId: "123",
			}

			storage.
				EXPECT().
				Find(id).
				Return(nil, nil)
			storage.EXPECT().
				Find(id).
				Return([]Model{
					{
						Identifier: id,
						Value:      val(1000),
					},
				}, nil).
				After(
					storage.
						EXPECT().
						Save(val(1000), id).
						Return(nil),
				)

			Expect(srv.Create(val(1000), id)).To(Succeed())
			err := srv.Create(val(1000), id)
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrAlreadyExist))

			err = srv.Create(val(1000), Identifier{Name: "test"})
			Expect(err).Should(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrIdIncomplete))
		})
		It("should find limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()
			storage := mockLimit.NewMockStorage(ctrl)
			id := Identifier{
				Name:     "limit1",
				Entity:   "user",
				EntityId: "123",
			}
			storage.
				EXPECT().
				Find(id).
				Return([]Model{{
					Identifier: id,
					Value:      val(100),
				}}, nil).
				AnyTimes()

			incId := Identifier{Entity: "user", EntityId: "123"}
			storage.
				EXPECT().
				Find(incId).
				Return([]Model{
					{
						Identifier: Identifier{
							Name:     "limit1",
							Entity:   incId.Entity,
							EntityId: incId.EntityId,
						},
						Value: val(100),
					},
					{
						Identifier: Identifier{
							Name:     "limit2",
							Entity:   incId.Entity,
							EntityId: incId.EntityId,
						},
						Value: val(500),
					},
				}, nil)

			srv := NewService(storage, NewFactory())

			found, err := srv.FindOne(id)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(found.Identifier()).To(Equal(id))

			available := found.Available()
			Expect(available.CurrencyAmount().Amount()).To(decEqual(dec(100)))

			limits, err := srv.Find(id)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(limits).To(HaveLen(1))
			Expect(limits[0].Identifier()).To(Equal(id))

			limits, err = srv.Find(incId)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(limits).To(HaveLen(2))
		})
		When("limit is not found", func() {
			It("should return error", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()
				storage := mockLimit.NewMockStorage(ctrl)
				storage.
					EXPECT().
					Find(any).
					Return(nil, ErrNotFound).
					AnyTimes()

				srv := NewService(storage, NewFactory())
				limits, err := srv.Find(Identifier{EntityId: "user"})
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrNotFound))
				Expect(limits).To(BeNil())

				limit, err := srv.FindOne(Identifier{EntityId: "user"})
				Expect(err).Should(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(ErrNotFound))
				Expect(limit).To(BeNil())

			})
		})
		It("should delete limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()
			storage := mockLimit.NewMockStorage(ctrl)

			id := Identifier{
				Name:     "some limit",
				Entity:   "balance",
				EntityId: "9876",
			}
			storage.
				EXPECT().
				Delete(id).
				Return(nil).
				AnyTimes()

			storage.
				EXPECT().
				Delete(Identifier{
					Entity:   "user",
					EntityId: "123",
				}).
				Return(nil)

			storage.
				EXPECT().
				Delete(Identifier{
					Name: "limit1",
				}).
				Return(nil)

			srv := NewService(storage, NewFactory())
			err := srv.DeleteOne(Identifier{
				EntityId: "111",
			})
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrIdIncomplete))

			Expect(srv.DeleteOne(id)).To(Succeed())
			Expect(srv.DeleteByEntity("user", "123")).To(Succeed())
			Expect(srv.DeleteByName("limit1")).To(Succeed())
		})
		It("should update limit", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()
			storage := mockLimit.NewMockStorage(ctrl)
			storage.
				EXPECT().
				Update(any, any).
				Return(nil).
				AnyTimes()

			srv := NewService(storage, NewFactory())
			err := srv.UpdateOne(val(99), Identifier{
				Name:   "limit1",
				Entity: "user",
			})
			Expect(err).Should(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(ErrIdIncomplete))
		})
		It("should ensure that transactional storage is supported", func() {
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()

			txStorage := mockLimit.NewMockTransactionalStorage(ctrl)
			txStorage.
				EXPECT().
				WrapContext(nil)

			srv := NewService(txStorage, NewFactory())
			srv.WrapContext(nil)
		})
	})

	Context("StorageGORM", func() {
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

		It("should save new limit", func() {
			id := Identifier{
				Name:     "name",
				Entity:   "entity",
				EntityId: "id",
			}
			value := val(10, "EUR")
			storage := NewStorageGORM(gdb)

			mock.
				ExpectExec("^INSERT INTO `limits`(.+) VALUES (.+)").
				WithArgs("EUR", "10", id.Name, id.Entity, id.EntityId).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := storage.Save(value, id)
			Expect(err).ShouldNot(HaveOccurred())

			mock.
				ExpectExec("^INSERT INTO `limits`(.+) VALUES (.+)").
				WithArgs(nil, nil, id.Name, id.Entity, id.EntityId).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = storage.Save(NoLimit(), id)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should update existing limit", func() {
			id := Identifier{
				Name:     "name",
				Entity:   "entity",
				EntityId: "id",
			}
			value := val(10, "EUR")
			storage := NewStorageGORM(gdb)

			mock.
				ExpectExec("^UPDATE `limits` SET .+").
				WithArgs("EUR", "10", id.Name, id.Entity, id.EntityId).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := storage.Update(value, id)
			Expect(err).ShouldNot(HaveOccurred())

			mock.
				ExpectExec("^UPDATE `limits` SET .+").
				WithArgs(nil, nil, id.Name, id.Entity, id.EntityId).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = storage.Update(NoLimit(), id)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should delete existing limit", func() {
			id := Identifier{
				Name:     "name",
				Entity:   "entity",
				EntityId: "id",
			}
			storage := NewStorageGORM(gdb)
			mock.ExpectBegin()
			mock.
				ExpectExec("^DELETE FROM `limits` WHERE .+").
				WithArgs(id.Name, id.Entity, id.EntityId).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			Expect(storage.Delete(id)).To(Succeed())
		})

		It("should find limit", func() {
			id := Identifier{
				Name:     "name",
				Entity:   "entity",
				EntityId: "id",
			}
			storage := NewStorageGORM(gdb)
			rows := sqlmock.NewRows([]string{
				"id",
				"currency_code",
				"amount",
				"name",
				"entity",
				"entity_id",
			})
			rows.AddRow(1, "EUR", "10", id.Name, id.Entity, id.EntityId)

			mock.
				ExpectQuery("^SELECT (.+) FROM `limits`  WHERE (.+)").
				WithArgs(id.Name, id.Entity, id.EntityId).
				WillReturnRows(rows)

			_, err := storage.Find(id)
			Expect(err).ToNot(HaveOccurred())

			mock.
				ExpectQuery("^SELECT (.+) FROM `limits`  WHERE (.+)").
				WithArgs(id.Entity, id.EntityId).
				WillReturnRows(rows)

			id.Name = ""
			_, err = storage.Find(id)

			Expect(err).ToNot(HaveOccurred())

			mock.
				ExpectQuery("^SELECT (.+) FROM `limits`  WHERE (.+)").
				WithArgs(id.Entity).
				WillReturnRows(rows)

			id.EntityId = ""
			_, err = storage.Find(id)

			Expect(err).ToNot(HaveOccurred())

			id.EntityId = "new"
			mock.
				ExpectQuery("^SELECT (.+) FROM `limits`  WHERE (.+)").
				WithArgs(id.EntityId).
				WillReturnRows(rows)

			id.Entity = ""
			_, err = storage.Find(id)

			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func val(available interface{}, currency ...string) Value {
	code := "EUR"
	if len(currency) > 0 {
		code = currency[0]
	}
	return Val(dec(available), code)
}

func amount(available interface{}, currency ...string) CurrencyAmount {
	code := "EUR"
	if len(currency) > 0 {
		code = currency[0]
	}
	return Amount(dec(available), code)
}

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

func decEqual(d interface{}) types.GomegaMatcher {
	return &decMatcher{
		Expected: dec(d),
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
