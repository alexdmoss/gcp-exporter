package client

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/tests"
)

func TestJwtConfigReader_read(t *testing.T) {
	tests.RunOnTempDir(t, "jwt-config-reader-read", func(t *testing.T, dir string) {
		tests.RunWithTempFile(t, dir, "service-account.json", func(t *testing.T, file string) {
			content := `
{
  "type": "service_account",
  "client_email": "service-account@example.com",
  "private_key_id": "key-id"
}
`

			ioutil.WriteFile(file, []byte(content), 0644)

			reader := NewJWTConfigReader(file)
			err := reader.Read()

			assert.NoError(t, err)
			require.NotNil(t, reader.JWTConfig())

			assert.Equal(t, "service-account@example.com", reader.JWTConfig().Email)
			assert.Equal(t, "key-id", reader.JWTConfig().PrivateKeyID)
		})
	})
}

func TestJwtConfigReader_read_invalidType(t *testing.T) {
	tests.RunOnTempDir(t, "jwt-config-reader-read", func(t *testing.T, dir string) {
		tests.RunWithTempFile(t, dir, "service-account.json", func(t *testing.T, file string) {
			content := `
{
  "type": "authorized_user",
  "client_secret": "secret",
  "client_id": "id"
}
`

			ioutil.WriteFile(file, []byte(content), 0644)

			reader := NewJWTConfigReader(file)
			err := reader.Read()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "could not parse service account JSON")
			assert.Contains(t, err.Error(), "'type' field is \"authorized_user\" (expected \"service_account\")")
		})
	})
}

func TestJwtConfigReader_read_invalidContent(t *testing.T) {
	tests.RunOnTempDir(t, "jwt-config-reader-read", func(t *testing.T, dir string) {
		tests.RunWithTempFile(t, dir, "service-account.json", func(t *testing.T, file string) {
			reader := NewJWTConfigReader(file)
			err := reader.Read()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "could not parse service account JSON")
			assert.Nil(t, reader.JWTConfig())
		})
	})
}

func TestJwtConfigReader_read_noFile(t *testing.T) {
	reader := NewJWTConfigReader("not-existing-service-account.json")
	err := reader.Read()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not read service account file")
	assert.Nil(t, reader.JWTConfig())
}
