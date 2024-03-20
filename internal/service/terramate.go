package service

import (
	"context"
	"fmt"
	"net/http"
	"github.com/rs/zerolog"
	cloud "github.com/terramate-io/terramate/cloud"
)

type SimpleTokenCredential struct {
	token string
}

func (c *SimpleTokenCredential) Token() (string, error) {
	return c.token, nil
}

func FetchUser(token string) (cloud.User, error) {
	client := cloud.Client{
		BaseURL:    cloud.BaseURL,
		Credential: &SimpleTokenCredential{token: token},
		HTTPClient: &http.Client{},
		Logger:     &zerolog.Logger{},
	}

	user, err := client.Users(context.Background())
	if err != nil {

		fmt.Printf("Failed to fetch user: %v\n", err)
		// Return the error directly, wrapped in fmt.Errorf to add additional context if needed.
		return cloud.User{}, fmt.Errorf("failed to fetch user: %w", err)
	}
	fmt.Printf("Fetched user: %+v\n", user)
	// Return the user object and nil for the error.
	return user, nil
}
