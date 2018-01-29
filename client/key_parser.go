package client

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
)

type KeyParserInterface interface {
	PrivateKey() *rsa.PrivateKey
	ParseKey() error
}

type KeyParser struct {
	privateKey *rsa.PrivateKey

	encodedKey []byte
}

func (kp *KeyParser) PrivateKey() *rsa.PrivateKey {
	return kp.privateKey
}

func (kp *KeyParser) ParseKey() error {
	logrus.Debugln("Parsing private key")

	privateKey := kp.encodedKey
	block, _ := pem.Decode(kp.encodedKey)
	if block != nil {
		privateKey = block.Bytes
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(privateKey)
	if err != nil {
		parsedKey, err = x509.ParsePKCS1PrivateKey(privateKey)
		if err != nil {
			return fmt.Errorf("private key should be a PEM or plain PKSC1 or PKCS8; parse error: %v", err)
		}
	}

	parsed, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return errors.New("private key is invalid")
	}

	kp.privateKey = parsed

	return nil
}

func NewKeyParser(encodedKey []byte) *KeyParser {
	return &KeyParser{
		encodedKey: encodedKey,
	}
}

var keyParserFactory = func(encodedKey []byte) KeyParserInterface {
	return NewKeyParser(encodedKey)
}
