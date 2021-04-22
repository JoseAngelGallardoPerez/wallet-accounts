// Code generated by MockGen. DO NOT EDIT.
// Source: ./factory.go

// Package mock_balance is a generated GoMock package.
package mock_balance

import (
	balance "github.com/Confialink/wallet-accounts/internal/modules/balance"
	gomock "github.com/golang/mock/gomock"
	gorm "github.com/jinzhu/gorm"
	reflect "reflect"
	time "time"
)

// MockAggregationFactory is a mock of AggregationFactory interface
type MockAggregationFactory struct {
	ctrl     *gomock.Controller
	recorder *MockAggregationFactoryMockRecorder
}

// MockAggregationFactoryMockRecorder is the mock recorder for MockAggregationFactory
type MockAggregationFactoryMockRecorder struct {
	mock *MockAggregationFactory
}

// NewMockAggregationFactory creates a new mock instance
func NewMockAggregationFactory(ctrl *gomock.Controller) *MockAggregationFactory {
	mock := &MockAggregationFactory{ctrl: ctrl}
	mock.recorder = &MockAggregationFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAggregationFactory) EXPECT() *MockAggregationFactoryMockRecorder {
	return m.recorder
}

// GeneralTotalByUserId mocks base method
func (m *MockAggregationFactory) GeneralTotalByUserId(userId string) (balance.Aggregator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GeneralTotalByUserId", userId)
	ret0, _ := ret[0].(balance.Aggregator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GeneralTotalByUserId indicates an expected call of GeneralTotalByUserId
func (mr *MockAggregationFactoryMockRecorder) GeneralTotalByUserId(userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GeneralTotalByUserId", reflect.TypeOf((*MockAggregationFactory)(nil).GeneralTotalByUserId), userId)
}

// TotalDebitedByUserIdPerPeriod mocks base method
func (m *MockAggregationFactory) TotalDebitedByUserIdPerPeriod(userId string, from, till time.Time) (balance.Aggregator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalDebitedByUserIdPerPeriod", userId, from, till)
	ret0, _ := ret[0].(balance.Aggregator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TotalDebitedByUserIdPerPeriod indicates an expected call of TotalDebitedByUserIdPerPeriod
func (mr *MockAggregationFactoryMockRecorder) TotalDebitedByUserIdPerPeriod(userId, from, till interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalDebitedByUserIdPerPeriod", reflect.TypeOf((*MockAggregationFactory)(nil).TotalDebitedByUserIdPerPeriod), userId, from, till)
}

// WrapContext mocks base method
func (m *MockAggregationFactory) WrapContext(db *gorm.DB) balance.AggregationFactory {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WrapContext", db)
	ret0, _ := ret[0].(balance.AggregationFactory)
	return ret0
}

// WrapContext indicates an expected call of WrapContext
func (mr *MockAggregationFactoryMockRecorder) WrapContext(db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WrapContext", reflect.TypeOf((*MockAggregationFactory)(nil).WrapContext), db)
}