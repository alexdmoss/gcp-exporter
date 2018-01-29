// Code generated by mockery v1.0.0

// This comment works around https://github.com/vektra/mockery/issues/155

package client

import mock "github.com/stretchr/testify/mock"
import oauth2 "golang.org/x/oauth2"

// MockTokenRequesterInterface is an autogenerated mock type for the TokenRequesterInterface type
type MockTokenRequesterInterface struct {
	mock.Mock
}

// RequestToken provides a mock function with given fields:
func (_m *MockTokenRequesterInterface) RequestToken() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Token provides a mock function with given fields:
func (_m *MockTokenRequesterInterface) Token() *oauth2.Token {
	ret := _m.Called()

	var r0 *oauth2.Token
	if rf, ok := ret.Get(0).(func() *oauth2.Token); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth2.Token)
		}
	}

	return r0
}
