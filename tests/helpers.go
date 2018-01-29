package tests

import (
	"bytes"
	"io/ioutil"
	"os"
	"syscall"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type ReadCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}

func RunOnHijackedLogrusOutput(t *testing.T, handler func(t *testing.T, output *bytes.Buffer)) {
	oldOutput := logrus.StandardLogger().Out
	defer func() { logrus.StandardLogger().Out = oldOutput }()

	buf := bytes.NewBuffer([]byte{})
	logrus.StandardLogger().Out = buf

	handler(t, buf)
}

func RunOnTempDir(t *testing.T, prefix string, handler func(t *testing.T, dir string)) {
	tempDir, err := ioutil.TempDir("", prefix)
	defer os.RemoveAll(tempDir)

	require.NoError(t, err)

	handler(t, tempDir)
}

func RunWithTempFile(t *testing.T, dir string, prefix string, handler func(t *testing.T, file string)) {
	file, err := ioutil.TempFile(dir, prefix)
	defer syscall.Unlink(file.Name())

	require.NoError(t, err)

	handler(t, file.Name())
}
