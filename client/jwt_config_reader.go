package client

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	"github.com/Sirupsen/logrus"
)

type JWTConfigReaderInterface interface {
	JWTConfig() *jwt.Config
	Read() error
}

type JWTConfigReader struct {
	jwtConfig              *jwt.Config
	serviceAccountFilePath string
}

func (jcr *JWTConfigReader) JWTConfig() *jwt.Config {
	return jcr.jwtConfig
}

func (jcr *JWTConfigReader) Read() error {
	logrus.Debugf("Reading JWT configuration from: %s", jcr.serviceAccountFilePath)

	sa, err := ioutil.ReadFile(jcr.serviceAccountFilePath)
	if err != nil {
		return fmt.Errorf("could not read service account file: %v", err)
	}

	logrus.Debugln("Parsing JWT configuration file")
	jcr.jwtConfig, err = google.JWTConfigFromJSON(sa)
	if err != nil {
		return fmt.Errorf("could not parse service account JSON: %v", err)
	}

	return nil
}

func NewJWTConfigReader(serviceAccountFilePath string) *JWTConfigReader {
	return &JWTConfigReader{
		serviceAccountFilePath: serviceAccountFilePath,
	}
}

var jwtConfigReaderFactory = func(serviceAccountFilePath string) JWTConfigReaderInterface {
	return NewJWTConfigReader(serviceAccountFilePath)
}
