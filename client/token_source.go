package client

import (
	"crypto/rsa"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"

	"github.com/Sirupsen/logrus"
)

type GCPServiceAccountTokenSource struct {
	serviceAccountFilePath string

	config     *jwt.Config
	privateKey *rsa.PrivateKey

	token *oauth2.Token
}

func (ts *GCPServiceAccountTokenSource) Token() (*oauth2.Token, error) {
	if ts.token != nil && ts.token.Valid() {
		logrus.Debugln("Re-using existing token")
		return ts.token, nil
	}

	logrus.Infoln("No token, or token expired; requesting new one")
	steps := []struct {
		command      func() error
		errorMessage string
	}{
		{ts.readJWTConfig, "could not read JWT configuration file"},
		{ts.parseKey, "could not parse RSA key"},
		{ts.requestForToken, "cloud not get Token"},
	}

	var err error
	for _, step := range steps {
		err = step.command()
		if err != nil {
			return nil, fmt.Errorf("failed on oauth2 token requesting: %s: %v", step.errorMessage, err)
		}
	}

	logrus.Infoln("New token saved")
	return ts.token, nil
}

func (ts *GCPServiceAccountTokenSource) readJWTConfig() error {
	reader := jwtConfigReaderFactory(ts.serviceAccountFilePath)
	err := reader.Read()
	if err != nil {
		return err
	}

	ts.config = reader.JWTConfig()

	return nil
}

func (ts *GCPServiceAccountTokenSource) parseKey() error {
	parser := keyParserFactory(ts.config.PrivateKey)
	err := parser.ParseKey()
	if err != nil {
		return err
	}

	ts.privateKey = parser.PrivateKey()

	return nil
}

func (ts *GCPServiceAccountTokenSource) requestForToken() error {
	tr := tokenRequesterFactory(ts.config, ts.privateKey)
	err := tr.RequestToken()
	if err != nil {
		return err
	}

	ts.token = tr.Token()

	return nil
}

func NewGCPServiceAccountTokenSource(serviceAccountFilePath string) *GCPServiceAccountTokenSource {
	return &GCPServiceAccountTokenSource{
		serviceAccountFilePath: serviceAccountFilePath,
	}
}
