// Code generated by mockery v1.0.0

// This comment works around https://github.com/vektra/mockery/issues/155

package client

import jws "golang.org/x/oauth2/jws"
import mock "github.com/stretchr/testify/mock"
import rsa "crypto/rsa"

// MockJWSEncoderInterface is an autogenerated mock type for the JWSEncoderInterface type
type MockJWSEncoderInterface struct {
	mock.Mock
}

// Encode provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockJWSEncoderInterface) Encode(_a0 *jws.Header, _a1 *jws.ClaimSet, _a2 *rsa.PrivateKey) (string, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 string
	if rf, ok := ret.Get(0).(func(*jws.Header, *jws.ClaimSet, *rsa.PrivateKey) string); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*jws.Header, *jws.ClaimSet, *rsa.PrivateKey) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
