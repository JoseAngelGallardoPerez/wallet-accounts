// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import (
	gin "github.com/gin-gonic/gin"
	mock "github.com/stretchr/testify/mock"
)

// AuthHandlerServiceInterface is an autogenerated mock type for the AuthHandlerServiceInterface type
type AuthHandlerServiceInterface struct {
	mock.Mock
}

// Can provides a mock function with given fields: context, action, resource
func (_m *AuthHandlerServiceInterface) Can(context *gin.Context, action string, resource string) bool {
	ret := _m.Called(context, action, resource)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*gin.Context, string, string) bool); ok {
		r0 = rf(context, action, resource)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}