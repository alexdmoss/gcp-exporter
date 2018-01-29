package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/tests"
)

func TestNew(t *testing.T) {
	tests.RunOnTempDir(t, "client-test-new", func(t *testing.T, dir string) {
		tests.RunWithTempFile(t, dir, "service-account.json", func(t *testing.T, file string) {
			_, err := New(file)
			assert.NoError(t, err)
		})
	})
}

func TestNew_NoFile(t *testing.T) {
	_, err := New("non-existing-file")
	assert.Error(t, err, "service account file non-existing-file doesn't exist: stat non-existing-file: no such file or directory")
}
