package config

import (
	"context"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/oauth2/v2"
)

// Getenvironment returns the environment for this instance
func Getenvironment() string {
	return os.Getenv("ENVIRONMENT")
}

// GetGkeNamespace returns the GKE Namespace for this instance
func GetGkeNamespace() string {
	return os.Getenv("GKE_NAMESPACE")
}

//GetGkePodName returns the GKE PodName for this instance
func GetGkePodName() string {
	return os.Getenv("GKE_PODNAME")
}

//GetTag returns the docker image tag / cloud build id
func GetTag() string {
	return os.Getenv("DOCKER_TAG")
}

//GetDeploymentName returns the deployment name for this instance.  This field only exists for multi-deployments
func GetDeploymentName() string {
	return os.Getenv("DEPLOYMENT_NAME")
}

// IsLocal returns true if this instance is running locally (meaning outside of GKE)
func IsLocal() bool {
	return GetGkeNamespace() == "" && GetGkePodName() == ""
}

// IsProd returns true if this instance is running on production
func IsProd() bool {
	return Getenvironment() == "production" || Getenvironment() == "prod"
}

// cache email
var email = ""

// GetCurrentLocalUserName returns the local user's name based on their google oauth email
func GetCurrentLocalUserName() string {
	if email != "" {
		return email
	}
	client, err := google.DefaultClient(context.Background(), oauth2.UserinfoEmailScope)
	if err != nil {
		panic(err)
	}
	service, err := oauth2.New(client)
	if err != nil {
		panic(err)
	}
	us := oauth2.NewUserinfoService(service)
	ui, err := us.Get().Do()
	if err != nil {
		panic(err)
	}
	email = ui.Email
	return email
}

// GetServiceAccount returns the google service account to use for this instance
// Deprecated: use GetGoogleServiceAccount instead
func GetServiceAccount() string {
	return GetGoogleServiceAccount()
}

// GetGoogleServiceAccount returns the google service account to use for this instance
func GetGoogleServiceAccount() string {
	if IsLocal() {
		return GetCurrentLocalUserName()
	}
	return os.Getenv("SERVICE_ACCOUNT")
}

// GetVendastaServiceAccount returns the vendasta service account to use for this instance
func GetVendastaServiceAccount() string {
	if IsLocal() {
		return GetCurrentLocalUserName()
	}
	return os.Getenv("VENDASTA_SERVICE_ACCOUNT")
}
