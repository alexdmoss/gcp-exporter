// Code generated by mockery v1.0.0

// This comment works around https://github.com/vektra/mockery/issues/155

package client

import http "net/http"
import mock "github.com/stretchr/testify/mock"

// MockHTTPRequestBuilderInterface is an autogenerated mock type for the HTTPRequestBuilderInterface type
type MockHTTPRequestBuilderInterface struct {
	mock.Mock
}

// NewRequest provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockHTTPRequestBuilderInterface) NewRequest(_a0 string, _a1 string, _a2 string) (*http.Request, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *http.Request
	if rf, ok := ret.Get(0).(func(string, string, string) *http.Request); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Request)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
