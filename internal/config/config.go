package config

import (
	"encoding/json"
	"os"
)

type CredentialData struct {
	IDToken string `json:"id_token"`
}

func LoadCredentials(filePath string) (string, error) {
	var creds CredentialData
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", err
	}
	return creds.IDToken, nil
}
