package logging

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/vendasta/gosdks/verrors"

	"cloud.google.com/go/logging"
	"github.com/mattheath/kala/bigflake"
	"github.com/vendasta/gosdks/statsd"
	"github.com/vendasta/gosdks/tracing"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	google_logging_type "google.golang.org/genproto/googleapis/logging/type"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

type CloudLoggingStrategy interface {
	// GetMetaDataLabels creates all the labels that get attach to the logs
	GetMetaDataLabels(namespace, resourceName, appName string) (map[string]string, error)
	// GetResourceType returns the value that will be set for `resource.type` examples types: "k8s_container", "cloud_function"
	GetResourceType() string
}

// This comes from how TraceID's type and stringer work
// TraceID is a [16]byte, and it's stringer outputs as base16 with 0's for padding
const missingTraceID = "00000000000000000000000000000000"

type cloudLogger struct {
	appLogger                 *logging.Logger
	commonLoggers             map[Stream]*logging.Logger
	config                    *config
	flake                     *bigflake.Bigflake
	normalizedPathFromRequest func(request *http.Request) string
}

// Compile time interface check
var _ Logger = &cloudLogger{}

func (l *cloudLogger) request(ctx context.Context, rl *requestData, stream Stream) {
	l.logRequest(ctx, rl, stream)
}

func (l *cloudLogger) Debugf(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Debug, f, a...)
}

func (l *cloudLogger) Infof(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Info, f, a...)
}

func (l *cloudLogger) Noticef(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Notice, f, a...)
}

func (l *cloudLogger) Warningf(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Warning, f, a...)
}

func (l *cloudLogger) Errorf(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Error, f, a...)
}

func (l *cloudLogger) Criticalf(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Critical, f, a...)
}

func (l *cloudLogger) Alertf(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Alert, f, a...)
}

func (l *cloudLogger) Emergencyf(ctx context.Context, f string, a ...interface{}) {
	l.log(ctx, logging.Emergency, f, a...)
}

// defaultPathFromRequest returns the URL path, but protects against nil requests/urls
func defaultPathFromRequest(request *http.Request) string {
	if request == nil || request.URL == nil {
		return ""
	}
	return request.URL.Path
}

func (l *cloudLogger) StackTrace(ctx context.Context, f string, a ...interface{}) {
	message := fmt.Sprintf(f, a...)
	stackTrace := debug.Stack()
	messageWithStack := fmt.Sprintf("%s:\n%s", message, string(stackTrace))
	l.Debugf(ctx, messageWithStack)
}

func (l *cloudLogger) Tag(ctx context.Context, key, value string) Logger {
	l.tag(ctx, key, value)
	return l
}

func (l *cloudLogger) getLogEntry(ctx context.Context, payload, filename string, lineNumber int64, severity logging.Severity) logging.Entry {
	cd, ok := taggedDataFromContext(ctx)
	if !ok {
		cd = &tagsData{
			tags: map[string]string{},
		}
	}
	var traceID string
	spanContext := tracing.FromContext(ctx).SpanContext()
	if spanContext.TraceID.String() != missingTraceID {
		traceID = fmt.Sprintf("projects/%s/traces/%s", l.config.ProjectID, spanContext.TraceID.String())
	}

	var source *logpb.LogEntrySourceLocation
	if filename != "" {
		source = &logpb.LogEntrySourceLocation{
			File: filename,
			Line: lineNumber,
		}
	}

	return logging.Entry{
		Timestamp: time.Now().UTC(),
		Severity:  severity,
		Labels:    cd.getLabels(),
		Payload:   payload,
		Operation: &logpb.LogEntryOperation{
			Producer: "gke-logger",
		},
		Trace:          traceID,
		SourceLocation: source,
	}
}

func (l *cloudLogger) log(ctx context.Context, severity logging.Severity, f string, a ...interface{}) {

	// If the severity of the log is less then the log inclusion configured, don't log
	if severity < logging.Severity(l.config.loggingInclusionLevel) {
		return
	}

	payload := f
	if len(a) > 0 {
		payload = fmt.Sprintf(f, a...)
	}

	var filename string
	var lineNumber int64
	if severity >= logging.Severity(l.config.filenameLoggingLevel) {
		_, file, line, ok := runtime.Caller(3)
		if ok {
			filename = file
			lineNumber = int64(line)
		}
	}

	if requestData, ok := requestDataFromContext(ctx); !ok {
		entry := l.getLogEntry(ctx, payload, filename, lineNumber, severity)
		if severity == logging.Emergency || severity == logging.Critical {
			// Synchronously log Emergency and Critical logs as the program is exiting soon
			_ = l.appLogger.LogSync(ctx, entry)
		} else {
			// Log message to cloud logging without bundling into an existing request log.
			l.appLogger.Log(entry)
		}
	} else {
		// We have an existing request associated to the provided context, add log message to existing
		// request data so we have our application logs bundled with our request log.
		requestData.logLine(payload, loggingSeverityToPB(severity), filename, lineNumber)
	}
}

func (l *cloudLogger) tag(ctx context.Context, key, value string) {
	requestData, ok := requestDataFromContext(ctx)

	if ok {
		// We have an existing request associated to the provided context, add tag to existing
		// request data
		requestData.addTag(key, value)
	} else {
		// otherwise try to use context data
		// this requires the context to have been created using NewTaggedContext
		contextData, ok := taggedDataFromContext(ctx)
		if ok {
			contextData.addTag(key, value)
		}
	}
}

func loggingSeverityToPB(s logging.Severity) google_logging_type.LogSeverity {
	return google_logging_type.LogSeverity(s)
}

func getStatusTags(status int) []string {
	statusStr := strconv.Itoa(status)
	series := statusStr[:1]
	return []string{
		fmt.Sprintf("status:%d", status),
		fmt.Sprintf("status_series:%sxx", series),
	}
}

func (l *cloudLogger) logRequest(ctx context.Context, rd *requestData, stream Stream) {
	tags := append(
		getStatusTags(rd.HTTPRequest.Status),
		fmt.Sprintf("namespace:%s", l.config.Namespace),
	)
	if rd.HTTPRequest.Request.URL != nil {
		normalizedPath := rd.HTTPRequest.Request.URL.Path
		if l.normalizedPathFromRequest != nil {
			normalizedPath = l.normalizedPathFromRequest(rd.HTTPRequest.Request)
		}
		tags = append(tags, fmt.Sprintf("path:%s", normalizedPath))
		l.Tag(ctx, "logging.normalized_path", normalizedPath)
	}

	if stream == "" || stream == Request {
		go func() {
			statsd.Histogram("gRPC.Latency", float64(rd.HTTPRequest.Latency.Nanoseconds()/1e6), tags, 1)
		}()
	}

	sev := logging.Info
	if rd.HTTPRequest.Request.URL != nil && rd.HTTPRequest.Request.URL.Path == "/healthz" {
		sev = logging.Debug
	}
	if rd.HTTPRequest.Status >= 400 && rd.HTTPRequest.Status < 500 {
		sev = logging.Warning
	}
	if rd.HTTPRequest.Status >= 500 {
		sev = logging.Error
	}

	// If the severity of the log is less then the log inclusion configured, don't log
	if sev < logging.Severity(l.config.loggingInclusionLevel) {
		return
	}

	setTraceData(ctx, rd, l.config.ProjectID)
	rd.Entry.Labels = rd.getLabels()
	rd.Entry.Severity = sev
	if logger, ok := l.commonLoggers[stream]; ok {
		logger.Log(*rd.Entry)
	}
	requestSuccessful := sev != logging.Error
	for _, line := range rd.lines {
		// Successful requests (non-500s) will never contain debug level lines or below
		if requestSuccessful && line.Severity <= google_logging_type.LogSeverity_DEBUG {
			continue
		}
		var source *logpb.LogEntrySourceLocation
		if line.Filename != "" {
			source = &logpb.LogEntrySourceLocation{
				File: line.Filename,
				Line: int64(line.LineNumber),
			}
		}
		l.appLogger.Log(logging.Entry{
			Timestamp:      line.Time,
			Severity:       logging.Severity(line.Severity),
			Payload:        line.LogMessage,
			Trace:          rd.Entry.Trace,
			Resource:       rd.Entry.Resource,
			SourceLocation: source,
		})
	}

}

// RequestID generates a new unique identifier for the current request, and should only be called
// once per request. Multiple calls to this method during the same request produce different request IDs.
func (l *cloudLogger) RequestID() string {
	f, err := l.requestID()
	if err != nil {
		Errorf(context.Background(), "Unable to use flake to generate a unique request id, returning empty string. Error: %s", err.Error())
		return ""
	}
	return f.String()
}

func (l *cloudLogger) requestID() (*bigflake.BigflakeId, error) {
	for {
		f, err := l.flake.Mint()
		if err == bigflake.ErrSequenceOverflow {
			time.Sleep(time.Millisecond)
			continue
		} else if err != nil {
			return nil, err
		}
		return f, nil
	}
}

func (l *cloudLogger) ActiveRequestId(ctx context.Context) string {
	rd, ok := requestDataFromContext(ctx)
	if !ok {
		return ""
	}
	return rd.requestID
}

func (l *cloudLogger) ActiveTraceId(ctx context.Context) string {
	return calculateTraceID(ctx, l.config)
}

// newCloudLogger creates a GKE logger that works with new logging settings in GKE clusters.
func newCloudLogger(config *config, client *logging.Client) (*cloudLogger, error) {
	if config.cloudLoggingStrategy == nil {
		return nil, verrors.New(verrors.InvalidArgument, "cloud logging strategy is required.")
	}

	labels, err := config.cloudLoggingStrategy.GetMetaDataLabels(config.Namespace, config.PodName, config.AppName)
	if err != nil {
		return nil, err
	}

	mr := &mrpb.MonitoredResource{Type: config.cloudLoggingStrategy.GetResourceType(), Labels: labels}

	flake, err := newFlake()
	if err != nil {
		return nil, fmt.Errorf("newCloudLogger: error initializing bigflake: %w", err)
	}

	client.OnError = func(err error) {
		fmt.Printf("Error flushing logs: %s", err.Error())
	}
	logger := cloudLogger{
		config:    config,
		flake:     flake,
		appLogger: client.Logger(config.AppName, logging.CommonResource(mr)),
		commonLoggers: map[Stream]*logging.Logger{
			Request:   client.Logger(string(Request), logging.CommonResource(mr), logging.ContextFunc(tracing.DisableLogTracing)),
			Pubsub:    client.Logger(string(Pubsub), logging.CommonResource(mr), logging.ContextFunc(tracing.DisableLogTracing)),
			Taskqueue: client.Logger(string(Taskqueue), logging.CommonResource(mr), logging.ContextFunc(tracing.DisableLogTracing)),
			Odin:      client.Logger(string(Odin), logging.CommonResource(mr), logging.ContextFunc(tracing.DisableLogTracing)),
			Goroutine: client.Logger(string(Goroutine), logging.CommonResource(mr), logging.ContextFunc(tracing.DisableLogTracing)),
		},
		normalizedPathFromRequest: defaultPathFromRequest,
	}

	if config.normalizedPathFromRequest != nil {
		logger.normalizedPathFromRequest = config.normalizedPathFromRequest
	}
	return &logger, nil
}

func newFlake() (*bigflake.Bigflake, error) {
	// workerID _must_ be unique between all instances of bigflake.New, in order to guarantee
	// unique requestID's. Since Initialize() is a singleton, if we just generate a random int
	// we can guarantee no two pods on the same GKE node share the same workerID.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// bigflake workerID must be less than 48 bits
	maxWorkerID := (1 << 48) - 1
	workerID := rng.Int63n(int64(maxWorkerID))
	flake, err := bigflake.New(uint64(workerID))
	if err != nil {
		return nil, err
	}
	return flake, nil
}
