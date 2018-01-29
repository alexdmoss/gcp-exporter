// Code generated by mockery v1.0.0

// This comment works around https://github.com/vektra/mockery/issues/155

package client

import mock "github.com/stretchr/testify/mock"
import rsa "crypto/rsa"

// MockKeyParserInterface is an autogenerated mock type for the KeyParserInterface type
type MockKeyParserInterface struct {
	mock.Mock
}

// ParseKey provides a mock function with given fields:
func (_m *MockKeyParserInterface) ParseKey() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PrivateKey provides a mock function with given fields:
func (_m *MockKeyParserInterface) PrivateKey() *rsa.PrivateKey {
	ret := _m.Called()

	var r0 *rsa.PrivateKey
	if rf, ok := ret.Get(0).(func() *rsa.PrivateKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rsa.PrivateKey)
		}
	}

	return r0
}
