package client

import (
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyParser_ParseKey(t *testing.T) {
	for _, fileName := range []string{"private.der", "private.pem", "private.pkcs8.der", "private.pkcs8.pem"} {
		t.Run(fileName, func(t *testing.T) {
			file := filepath.Join("testdata", fileName)

			encodedKey, err := ioutil.ReadFile(file)
			require.NoError(t, err)

			parser := NewKeyParser(encodedKey)
			err = parser.ParseKey()

			assert.NoError(t, err)
			assert.IsType(t, &rsa.PrivateKey{}, parser.PrivateKey())

			encrypedFile := filepath.Join("testdata", "encrypted")
			encryptedData, err := ioutil.ReadFile(encrypedFile)
			require.NoError(t, err)

			decryptedData, err := parser.PrivateKey().Decrypt(rand.Reader, encryptedData, nil)
			assert.NoError(t, err)
			assert.Equal(t, "decrypted content", string(decryptedData))
		})
	}
}

func TestKeyParser_ParseKey_InvalidKey(t *testing.T) {
	encodedKey := []byte{'a', 'b', 'c'}

	parser := NewKeyParser(encodedKey)
	err := parser.ParseKey()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "private key should be a PEM or plain PKSC1 or PKCS8; parse error")
}
