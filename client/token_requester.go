package client

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
	"golang.org/x/oauth2/jwt"

	"github.com/Sirupsen/logrus"
)

const (
	TokenRequestURL = "https://www.googleapis.com/oauth2/v4/token"
	ClaimScope      = "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/compute.readonly"
	ClaimAUD        = "https://www.googleapis.com/oauth2/v4/token"
	JWTGrantType    = "urn:ietf:params:oauth:grant-type:jwt-bearer"
)

type HTTPClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

type JWSEncoderInterface interface {
	Encode(*jws.Header, *jws.ClaimSet, *rsa.PrivateKey) (string, error)
}

type JWSEncoder struct{}

func (je *JWSEncoder) Encode(header *jws.Header, c *jws.ClaimSet, key *rsa.PrivateKey) (string, error) {
	return jws.Encode(header, c, key)
}

type HTTPRequestBuilderInterface interface {
	NewRequest(string, string, string) (*http.Request, error)
}

type HTTPRequestBuilder struct{}

func (rb *HTTPRequestBuilder) NewRequest(method, url string, body string) (*http.Request, error) {
	return http.NewRequest(method, url, bytes.NewBufferString(body))
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (tresp *tokenResponse) isValid() bool {
	return tresp.AccessToken != "" &&
		tresp.TokenType != "" &&
		tresp.ExpiresIn != 0
}

type TokenRequesterInterface interface {
	Token() *oauth2.Token
	RequestToken() error
}

type TokenRequester struct {
	token *oauth2.Token

	config     *jwt.Config
	privateKey *rsa.PrivateKey

	client HTTPClientInterface
	jwsEnc JWSEncoderInterface
	httpRB HTTPRequestBuilderInterface
}

func (tr *TokenRequester) Token() *oauth2.Token {
	return tr.token
}

func (tr *TokenRequester) RequestToken() error {
	logrus.Debugln("Requesting new oAuth2 token")

	request, err := tr.prepareRequest()
	if err != nil {
		return err
	}

	response, err := tr.client.Do(request)
	if err != nil {
		return fmt.Errorf("error during HTTP Request: %v", err)
	}

	defer response.Body.Close()
	tokenResponse, err := tr.parseResponse(response)
	if err != nil {
		return err
	}

	tr.token = &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}

	return nil
}

func (tr *TokenRequester) prepareRequest() (*http.Request, error) {
	logrus.Debugln("Preparing request")
	jwsHeader := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
	}

	iat := time.Now()
	exp := iat.Add(time.Hour)

	jwsClaim := &jws.ClaimSet{
		Iss:   tr.config.Email,
		Scope: ClaimScope,
		Aud:   ClaimAUD,
		Exp:   exp.Unix(),
		Iat:   iat.Unix(),
	}

	logrus.Debugln("Encoding JWT assertion")
	jwsAssertion, err := tr.jwsEnc.Encode(jwsHeader, jwsClaim, tr.privateKey)
	if err != nil {
		return nil, fmt.Errorf("could not encode JWT: %v", err)
	}

	logrus.Debugln("Creating request")
	body := fmt.Sprintf("grant_type=%s&assertion=%s", url.PathEscape(JWTGrantType), jwsAssertion)
	tokenRequest, err := tr.httpRB.NewRequest(http.MethodPost, TokenRequestURL, body)
	if err != nil {
		return nil, fmt.Errorf("could not prepare HTTP Request: %v", err)
	}

	tokenRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return tokenRequest, nil
}

func (tr *TokenRequester) parseResponse(response *http.Response) (tokenResponse, error) {
	logrus.Debugln("Reading response body")

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("error while reading response body: %v", err)
	}

	logrus.Debugln("Parsing response body")

	var tokenResp tokenResponse
	err = json.Unmarshal(responseBody, &tokenResp)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("error while parsing response body: %v", err)
	}

	if !tokenResp.isValid() {
		return tokenResponse{}, fmt.Errorf("error while parsing response body: expected values are empty")
	}

	logrus.WithFields(logrus.Fields{
		"TokenType": tokenResp.TokenType,
		"ExpiresIn": tokenResp.ExpiresIn,
	}).Info("Received new token")

	return tokenResp, nil
}

func NewTokenRequester(config *jwt.Config, privateKey *rsa.PrivateKey) *TokenRequester {
	return &TokenRequester{
		config:     config,
		privateKey: privateKey,
		client:     http.DefaultClient,
		jwsEnc:     &JWSEncoder{},
		httpRB:     &HTTPRequestBuilder{},
	}
}

var tokenRequesterFactory = func(config *jwt.Config, privateKey *rsa.PrivateKey) TokenRequesterInterface {
	return NewTokenRequester(config, privateKey)
}
