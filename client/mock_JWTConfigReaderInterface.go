// Code generated by mockery v1.0.0

// This comment works around https://github.com/vektra/mockery/issues/155

package client

import jwt "golang.org/x/oauth2/jwt"
import mock "github.com/stretchr/testify/mock"

// MockJWTConfigReaderInterface is an autogenerated mock type for the JWTConfigReaderInterface type
type MockJWTConfigReaderInterface struct {
	mock.Mock
}

// JWTConfig provides a mock function with given fields:
func (_m *MockJWTConfigReaderInterface) JWTConfig() *jwt.Config {
	ret := _m.Called()

	var r0 *jwt.Config
	if rf, ok := ret.Get(0).(func() *jwt.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*jwt.Config)
		}
	}

	return r0
}

// Read provides a mock function with given fields:
func (_m *MockJWTConfigReaderInterface) Read() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}