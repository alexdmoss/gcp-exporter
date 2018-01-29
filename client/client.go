package client

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func newClient(ctx context.Context, serviceAccountFilePath string) (*http.Client, error) {
	_, err := os.Stat(serviceAccountFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("service account file %s doesn't exist: %v", serviceAccountFilePath, err)
	}

	if os.IsPermission(err) {
		return nil, fmt.Errorf("service account file %s cannot be read, because of permission problems: %v", serviceAccountFilePath, err)
	}

	return oauth2.NewClient(ctx, NewGCPServiceAccountTokenSource(serviceAccountFilePath)), nil
}

func New(serviceAccountFilePath string) (*http.Client, error) {
	return newClient(context.Background(), serviceAccountFilePath)
}
