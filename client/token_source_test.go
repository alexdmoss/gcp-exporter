package client

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/tests"
)

func runWithGCPServiceAccountTokenSource(t *testing.T,
	expectedEncodedKey interface{},
	expectedJWTConfig interface{},
	expectedPrivateKey interface{},
	handler func(t *testing.T, ts *GCPServiceAccountTokenSource, jwtcr *MockJWTConfigReaderInterface, kp *MockKeyParserInterface, tr *MockTokenRequesterInterface),
) {
	serviceAccountFile := "fake-file"

	jwtcr := &MockJWTConfigReaderInterface{}
	jwtConfigReaderFactory = func(serviceAccountFilePath string) JWTConfigReaderInterface {
		assert.Equal(t, serviceAccountFile, serviceAccountFilePath)

		return jwtcr
	}

	kp := &MockKeyParserInterface{}
	keyParserFactory = func(encodedKey []byte) KeyParserInterface {
		assert.Equal(t, expectedEncodedKey, encodedKey)

		return kp
	}

	tr := &MockTokenRequesterInterface{}
	tokenRequesterFactory = func(config *jwt.Config, privateKey *rsa.PrivateKey) TokenRequesterInterface {
		assert.Equal(t, expectedJWTConfig, config)
		assert.Equal(t, expectedPrivateKey, privateKey)

		return tr
	}

	ts := NewGCPServiceAccountTokenSource(serviceAccountFile)

	handler(t, ts, jwtcr, kp, tr)
}

func TestNewGCPServiceAccountTokenSource_Token_ValidToken(t *testing.T) {
	runWithGCPServiceAccountTokenSource(t, nil, nil, nil, func(t *testing.T, ts *GCPServiceAccountTokenSource, jwtcr *MockJWTConfigReaderInterface, kp *MockKeyParserInterface, tr *MockTokenRequesterInterface) {
		ts.token = &oauth2.Token{
			AccessToken: "fake-token",
			Expiry:      time.Now().Add(60 * time.Second),
		}

		token, err := ts.Token()

		assert.NoError(t, err)
		assert.Equal(t, "fake-token", token.AccessToken)

		jwtcr.AssertNotCalled(t, "Read")
		jwtcr.AssertNotCalled(t, "JWTConfig")

		kp.AssertNotCalled(t, "ParseKey")
		kp.AssertNotCalled(t, "PrivateKey")

		tr.AssertNotCalled(t, "RequestToken")
		tr.AssertNotCalled(t, "Token")
	})
}

func testRequestingToken(t *testing.T, existingToken *oauth2.Token) {
	encodedKey := []byte("private-key")
	config := &jwt.Config{PrivateKey: encodedKey}
	privateKey := &rsa.PrivateKey{}

	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		runWithGCPServiceAccountTokenSource(t, encodedKey, config, privateKey, func(t *testing.T, ts *GCPServiceAccountTokenSource, jwtcr *MockJWTConfigReaderInterface, kp *MockKeyParserInterface, tr *MockTokenRequesterInterface) {
			ts.token = existingToken

			newToken := &oauth2.Token{}

			jwtcr.On("Read").Return(nil)
			jwtcr.On("JWTConfig").Return(config)

			kp.On("ParseKey").Return(nil)
			kp.On("PrivateKey").Return(privateKey)

			tr.On("RequestToken").Return(nil)
			tr.On("Token").Return(newToken)

			token, err := ts.Token()

			assert.NoError(t, err)
			assert.Equal(t, newToken, token)
			assert.Contains(t, output.String(), "No token, or token expired; requesting new one")
			assert.Contains(t, output.String(), "New token saved")
		})
	})
}

func TestNewGCPServiceAccountTokenSource_Token(t *testing.T) {
	examples := map[string]*oauth2.Token{
		"empty_token":   {AccessToken: "", Expiry: time.Now().Add(60 * time.Second)},
		"expired_token": {AccessToken: "fake-token", Expiry: time.Now().Add(-60 * time.Second)},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			testRequestingToken(t, example)
		})
	}
}

func TestNewGCPServiceAccountTokenSource_Token_failureOnSteps(t *testing.T) {
	runWithGCPServiceAccountTokenSource(t, nil, nil, nil, func(t *testing.T, ts *GCPServiceAccountTokenSource, jwtcr *MockJWTConfigReaderInterface, kp *MockKeyParserInterface, tr *MockTokenRequesterInterface) {
		ts.token = &oauth2.Token{
			AccessToken: "",
		}

		jwtcr.On("Read").Return(fmt.Errorf("fake-jwt-config-reader-error"))

		token, err := ts.Token()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed on oauth2 token requesting: could not read JWT configuration file: fake-jwt-config-reader-error")

		assert.Nil(t, token)

		jwtcr.AssertNotCalled(t, "JWTConfig")

		kp.AssertNotCalled(t, "ParseKey")
		kp.AssertNotCalled(t, "PrivateKey")

		tr.AssertNotCalled(t, "RequestToken")
		tr.AssertNotCalled(t, "Token")
	})
}
