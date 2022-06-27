package config

import (
	"github.com/vendasta/gosdks/verrors"
	"io/ioutil"
	"os"
)

// environment variable names
const (
	applicationCredentialsJSON = "VENDASTA_APPLICATION_CREDENTIALS_JSON"
	applicationCredentials     = "VENDASTA_APPLICATION_CREDENTIALS"
	serviceAccountEmail        = "VENDASTA_SERVICE_ACCOUNT"
	publicKeyID                = "VENDASTA_PUBLIC_KEY_ID"
)

// VendastaApplicationCredentialsJSON returns the vendasta application credentials JSON file
func VendastaApplicationCredentialsJSON() string {
	return os.Getenv(applicationCredentialsJSON)
}

// VendastaApplicationCredentials returns the vendasta application credentials
func VendastaApplicationCredentials() (string, error) {
	bytes, err := ioutil.ReadFile(os.Getenv(applicationCredentials))
	if err != nil {
		return "", verrors.New(
			verrors.Internal,
			"Could not read %s environment variable: %s", applicationCredentials, err.Error(),
		)
	}
	return string(bytes), nil
}

// ServiceAccountEmail returns service account email for the current microservice
func ServiceAccountEmail() string {
	return os.Getenv(serviceAccountEmail)
}

// PublicKeyID returns the public key for the current microservice
func PublicKeyID() (string, error) {
	bytes, err := ioutil.ReadFile(os.Getenv(publicKeyID))
	if err != nil {
		return "", verrors.New(
			verrors.Internal,
			"Unable to read environment variable %s: %s", publicKeyID, err.Error(),
		)
	}
	return string(bytes), nil
}
