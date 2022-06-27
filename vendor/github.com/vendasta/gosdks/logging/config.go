package logging

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	gce_metadata "cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

var configValue *config
var mut sync.Mutex

type Level logging.Severity

const (
	LevelDebug     = Level(logging.Debug)
	LevelInfo      = Level(logging.Info)
	LevelWarning   = Level(logging.Warning)
	LevelError     = Level(logging.Error)
	LevelCritical  = Level(logging.Critical)
	LevelAlert     = Level(logging.Alert)
	LevelEmergency = Level(logging.Emergency)
)

type config struct {
	ProjectID             string
	Namespace             string
	PodName               string
	AppName               string
	cloudLogging          bool
	cloudLoggingStrategy  CloudLoggingStrategy
	filenameLoggingLevel  Level
	loggingInclusionLevel Level
	resourceType          string

	normalizedPathFromRequest func(request *http.Request) string
}

func (c *config) BuildHeader(name string) string {
	if c == nil {
		return "local"
	}
	return fmt.Sprintf("x-%s-%s", strings.ToLower(c.AppName), name)
}

// Initialize must be called on app startup and must be done before any logging statements have been issued.
func Initialize(gkeNamespace, podName, appName string, opts ...LoggerOption) error {
	if appName == "" {
		return errors.New("appName must be supplied")
	}
	mut.Lock()
	defer mut.Unlock()

	if configValue != nil {
		return nil
	}
	projectID := appName + "-local"
	if gce_metadata.OnGCE() {
		var err error
		projectID, err = gce_metadata.ProjectID()
		if err != nil {
			return err
		}
	}

	configValue = &config{
		Namespace:             gkeNamespace,
		PodName:               podName,
		AppName:               appName,
		ProjectID:             projectID,
		cloudLogging:          true,
		filenameLoggingLevel:  LevelError,
		loggingInclusionLevel: getLogLevel(),
	}

	for _, opt := range opts {
		opt(configValue)
	}

	ctx := context.Background()
	if gce_metadata.OnGCE() && configValue.cloudLogging {
		if gkeNamespace == "" || podName == "" {
			return errors.New("gkeNamespace and podName must be supplied")
		}
		//If no logging strategy was provided default to GKE
		if configValue.cloudLoggingStrategy == nil {
			configValue.cloudLoggingStrategy = gkeCloudLoggingStrategy{}
			// If we are running as a cloud function default to cloud function instead
			if os.Getenv("K_SERVICE") != "" { // Running as a cloud function
				configValue.cloudLoggingStrategy = cloudFunctionCloudLoggingStrategy{}
			}
		}

		client, err := logging.NewClient(ctx, projectID, option.WithGRPCConnectionPool(5))
		if err != nil {
			return err
		}
		loggerInstance, err = newCloudLogger(configValue, client)
		if err != nil {
			return err
		}
	} else {
		loggerInstance, _ = newStdErrLogger(configValue)
	}
	return nil
}

func getLogLevel() Level {
	logLevel := os.Getenv("LOG_INCLUSION_LEVEL")
	if logLevel == "" {
		return LevelInfo
	}

	logLevel = strings.ToLower(logLevel)
	switch logLevel {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warning":
		return LevelWarning
	case "error":
		return LevelError
	case "critical":
		return LevelCritical
	case "alert":
		return LevelAlert
	case "emergency":
		return LevelEmergency
	default:
		return LevelInfo
	}
}
