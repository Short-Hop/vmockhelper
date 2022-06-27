package logging

import (
	"net/http"
)

// LoggerOption augments the behavior of the cloudLogger
type LoggerOption func(*config)

// PathNormalizer sets the normalization strategy for path tags, for logging/datadog
// To normalize path from the http.Request, use `NormalizedPathFromRequest`
func PathNormalizer(normalizer func(string) string) LoggerOption {
	return func(l *config) {
		l.normalizedPathFromRequest = func(request *http.Request) string {
			return normalizer(request.URL.Path)
		}
	}
}

// NormalizedPathFromRequest sets the normalization strategy for path tags from request, for logging/datadog
// Example reasons:
//	vStatic paths like /some-client.123456.prod/main.123456.js => /some-client/main.js
//  REST paths handling method and path params like `GET /order/{orderID}/status`
func NormalizedPathFromRequest(normalizedPathFromRequest func(*http.Request) string) LoggerOption {
	return func(l *config) {
		l.normalizedPathFromRequest = normalizedPathFromRequest
	}
}

func LocalLoggingOnly() LoggerOption {
	return func(l *config) {
		l.cloudLogging = false
		l.filenameLoggingLevel = LevelDebug
	}
}

// If you are LocalLoggingOnly and FilenameLogLevel, add the FilenameLogLevel option after LocalLoggingOnly.
func FilenameLogLevel(l Level) LoggerOption {
	return func(c *config) {
		c.filenameLoggingLevel = l
	}
}

// WithCloudLoggingStrategy allows you to provide your own cloud logging strategy
func WithCloudLoggingStrategy(strat CloudLoggingStrategy) LoggerOption {
	return func(l *config) {
		l.cloudLoggingStrategy = strat
	}
}
