package client

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"golang.org/x/oauth2/jwt"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/tests"
)

func getTokenRequester(t *testing.T) (*TokenRequester, *jwt.Config, *rsa.PrivateKey) {
	config := &jwt.Config{
		Email: "service-account@example.com",
	}

	privateKeyFile := filepath.Join("testdata", "private.der")
	privateKeyBytes, err := ioutil.ReadFile(privateKeyFile)
	require.NoError(t, err)

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	require.NoError(t, err)
	require.IsType(t, &rsa.PrivateKey{}, privateKey)

	return NewTokenRequester(config, privateKey), config, privateKey
}

func TestTokenRequester_RequestToken_jwsEncodingFailure(t *testing.T) {
	tr, _, privateKey := getTokenRequester(t)

	jwsEnc := &MockJWSEncoderInterface{}
	jwsEnc.On("Encode", mock.Anything, mock.Anything, privateKey).
		Return("", fmt.Errorf("fake-jwsenc-error")).Once()
	defer jwsEnc.AssertExpectations(t)
	tr.jwsEnc = jwsEnc

	err := tr.RequestToken()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not encode JWT: fake-jwsenc-error")
}

func runWithJwsEnc(t *testing.T, tr *TokenRequester, privateKey *rsa.PrivateKey, handler func(t *testing.T, jwsEnc *MockJWSEncoderInterface, jwsAssertion string)) {
	jwsAssertion := "fake-data"
	jwsEnc := &MockJWSEncoderInterface{}
	jwsEnc.On("Encode", mock.Anything, mock.Anything, privateKey).Return(jwsAssertion, nil).Once()
	defer jwsEnc.AssertExpectations(t)
	tr.jwsEnc = jwsEnc

	handler(t, jwsEnc, jwsAssertion)
}

func TestTokenRequester_RequestToken_httpRequestBuildingFailure(t *testing.T) {
	tr, _, privateKey := getTokenRequester(t)

	runWithJwsEnc(t, tr, privateKey, func(t *testing.T, jwsEnc *MockJWSEncoderInterface, jwsAssertion string) {
		httpRB := &MockHTTPRequestBuilderInterface{}
		httpRB.On("NewRequest", http.MethodPost, "https://www.googleapis.com/oauth2/v4/token", fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%s", jwsAssertion)).
			Return(nil, fmt.Errorf("fake-httprb-error")).Once()
		defer httpRB.AssertExpectations(t)
		tr.httpRB = httpRB

		err := tr.RequestToken()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not prepare HTTP Request: fake-httprb-error")
	})
}

func runWithHTTPRB(t *testing.T, tr *TokenRequester, jwsAssertion string, handler func(t *testing.T, httpRB *MockHTTPRequestBuilderInterface, request *http.Request)) {
	request, err := http.NewRequest(http.MethodPost, "http://localhost:1234", nil)
	require.NoError(t, err)

	httpRB := &MockHTTPRequestBuilderInterface{}
	httpRB.On("NewRequest", http.MethodPost, "https://www.googleapis.com/oauth2/v4/token", fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%s", jwsAssertion)).
		Return(request, nil).Once()
	defer httpRB.AssertExpectations(t)
	tr.httpRB = httpRB

	handler(t, httpRB, request)
}

func TestTokenRequester_RequestToken_httpRequestFailure(t *testing.T) {
	tr, _, privateKey := getTokenRequester(t)

	runWithJwsEnc(t, tr, privateKey, func(t *testing.T, jwsEnc *MockJWSEncoderInterface, jwsAssertion string) {
		runWithHTTPRB(t, tr, jwsAssertion, func(t *testing.T, httpRB *MockHTTPRequestBuilderInterface, request *http.Request) {
			httpClient := &MockHTTPClientInterface{}
			httpClient.On("Do", request).Return(nil, fmt.Errorf("fake-http-error")).Once()
			defer httpClient.AssertExpectations(t)
			tr.client = httpClient

			err := tr.RequestToken()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "error during HTTP Request: fake-http-error")
		})
	})
}

func runWithHTTPClient(t *testing.T, tr *TokenRequester, request *http.Request, handler func(t *testing.T, httpClient *MockHTTPClientInterface, response *http.Response)) {
	response := &http.Response{}

	httpClient := &MockHTTPClientInterface{}
	httpClient.On("Do", request).Return(response, nil).Once()
	defer httpClient.AssertExpectations(t)
	tr.client = httpClient

	handler(t, httpClient, response)
}

func TestTokenRequester_RequestToken_httpResponseBodyReadFailure(t *testing.T) {
	tr, _, privateKey := getTokenRequester(t)

	runWithJwsEnc(t, tr, privateKey, func(t *testing.T, jwsEnc *MockJWSEncoderInterface, jwsAssertion string) {
		runWithHTTPRB(t, tr, jwsAssertion, func(t *testing.T, httpRB *MockHTTPRequestBuilderInterface, request *http.Request) {
			runWithHTTPClient(t, tr, request, func(t *testing.T, httpClient *MockHTTPClientInterface, response *http.Response) {
				body := &tests.MockReadCloser{}
				body.On("Read", mock.Anything).Return(0, fmt.Errorf("fake-body-read-error")).Once()
				body.On("Close").Return(nil).Once()
				defer body.AssertExpectations(t)
				response.Body = body

				err := tr.RequestToken()

				require.Error(t, err)
				assert.Contains(t, err.Error(), "error while reading response body: fake-body-read-error")
			})
		})
	})
}

func TestTokenRequester_RequestToken_httpResponseInvalidContent(t *testing.T) {
	tr, _, privateKey := getTokenRequester(t)

	examples := []struct {
		content  string
		errorMsg string
	}{
		{content: "<908xkqjkl", errorMsg: "invalid character"},
		{content: `{"test": "test"}`, errorMsg: "expected values are empty"},
	}

	for exampleID, example := range examples {
		t.Run(strconv.Itoa(exampleID), func(t *testing.T) {
			runWithJwsEnc(t, tr, privateKey, func(t *testing.T, jwsEnc *MockJWSEncoderInterface, jwsAssertion string) {
				runWithHTTPRB(t, tr, jwsAssertion, func(t *testing.T, httpRB *MockHTTPRequestBuilderInterface, request *http.Request) {
					runWithHTTPClient(t, tr, request, func(t *testing.T, httpClient *MockHTTPClientInterface, response *http.Response) {
						response.Body = ioutil.NopCloser(bytes.NewBufferString(example.content))

						err := tr.RequestToken()

						require.Error(t, err)
						assert.Contains(t, err.Error(), "error while parsing response body:")
						assert.Contains(t, err.Error(), example.errorMsg)
					})
				})
			})
		})
	}
}

func TestTokenRequester_RequestToken(t *testing.T) {
	tr, _, privateKey := getTokenRequester(t)

	runWithJwsEnc(t, tr, privateKey, func(t *testing.T, jwsEnc *MockJWSEncoderInterface, jwsAssertion string) {
		runWithHTTPRB(t, tr, jwsAssertion, func(t *testing.T, httpRB *MockHTTPRequestBuilderInterface, request *http.Request) {
			runWithHTTPClient(t, tr, request, func(t *testing.T, httpClient *MockHTTPClientInterface, response *http.Response) {
				tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
					content := `
{
  "access_token": "token",
  "token_type": "type",
  "expires_in": 20
}
`
					response.Body = ioutil.NopCloser(bytes.NewBufferString(content))

					err := tr.RequestToken()

					assert.NoError(t, err)
					assert.Equal(t, "token", tr.Token().AccessToken)
					assert.Equal(t, "type", tr.Token().TokenType)
					assert.True(t, time.Now().Add(10*time.Second).Unix() < tr.Token().Expiry.Unix())

					assert.Contains(t, output.String(), "Received new token")
					assert.Contains(t, output.String(), "ExpiresIn=20")
					assert.Contains(t, output.String(), "TokenType=type")
				})
			})
		})
	})
}
