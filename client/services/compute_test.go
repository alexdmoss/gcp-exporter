package services

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRoundTripper struct {
	t *testing.T
}

func (frt *fakeRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	unauthorizedResponseBody := `
{
  "error": {
    "errors": [
      {
		"domain": "global",
		"reason": "required",
		"message": "Login Required",
		"locationType": "header",
		"location": "Authorization"
      }
	],
	"code": 401,
	"message": "Login Required"
  }
}
`

	bodyBuffer := bytes.NewBufferString(unauthorizedResponseBody)
	body := ioutil.NopCloser(bodyBuffer)

	resp := &http.Response{
		Status:     "401 Unauthorized",
		StatusCode: 401,
		Body:       body,
	}

	return resp, nil
}

func getFakeClient(t *testing.T) *http.Client {
	rt := &fakeRoundTripper{t: t}
	client := http.DefaultClient
	client.Transport = rt

	return client
}

func TestComputeService_ListInstances_notAuthorized(t *testing.T) {
	c, err := NewComputeService(getFakeClient(t))
	assert.NoError(t, err)

	instancesList, err := c.ListInstances("fake-project", "fake-zone", 10)

	assert.Empty(t, instancesList)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error 401: Login Required")
}

func TestComputeService_ListInstances_notInitialized(t *testing.T) {
	c := &ComputeService{}
	instancesList, err := c.ListInstances("fake-project", "fake-zone", 10)

	assert.Empty(t, instancesList)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestComputeService_GetRegion(t *testing.T) {
	c, err := NewComputeService(getFakeClient(t))
	assert.NoError(t, err)

	reg, err := c.GetRegion("fake-project", "fake-region")

	assert.Nil(t, reg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error 401: Login Required")
}

func TestComputeService_GetRegion_notInitialized(t *testing.T) {
	c := &ComputeService{}
	reg, err := c.GetRegion("fake-project", "fake-region")

	assert.Nil(t, reg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}
