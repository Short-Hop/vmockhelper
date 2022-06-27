package logging

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"github.com/vendasta/gosdks/tracing"
	google_logging_type "google.golang.org/genproto/googleapis/logging/type"
)

type requestDataKey struct{}

type requestData struct {
	startTime time.Time
	requestID string
	mu        sync.RWMutex
	*logging.Entry
	lines []*logLine

	// Additional labels to add to the GKE request
	tags map[string]string

	// common labels will override tags and should only be filled with labels common to all requests
	commonLabels map[string]string
}

type logLine struct {
	time.Time
	Severity   google_logging_type.LogSeverity
	LogMessage string
	Filename   string
	LineNumber int64
}

func (rd *requestData) logLine(message string, severity google_logging_type.LogSeverity, filename string, line int64) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	rd.lines = append(rd.lines, &logLine{
		Time:       time.Now().UTC(),
		Severity:   severity,
		LogMessage: message,
		Filename:   filename,
		LineNumber: line,
	})
}

func (rd *requestData) addTag(key, value string) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	rd.tags[key] = value
}

func (rd *requestData) getLabels() map[string]string {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	r := map[string]string{}

	for k, v := range rd.tags {
		r[k] = v
	}

	for k, v := range rd.commonLabels {
		r[k] = v
	}

	return r
}

func newRequest(ctx context.Context, requestID string) (context.Context, *requestData) {
	td, ok := taggedDataFromContext(ctx)
	var labels map[string]string
	if !ok {
		labels = map[string]string{}
	} else {
		labels = td.getLabels()
	}
	projectID, err := metadata.ProjectID()
	if err != nil {
		projectID = "repcore-prod"
	}

	rd := &requestData{
		startTime: time.Now().UTC(),
		requestID: requestID,
		Entry: &logging.Entry{
			// this request will be replaced with the actual request in http interceptors
			HTTPRequest: &logging.HTTPRequest{
				Request: &http.Request{},
			},
		},
		tags: labels,
		commonLabels: map[string]string{
			"module_id":  "default",
			"version_id": "default",
			"project_id": projectID,
			"request_id": requestID,
		},
	}
	setTraceData(ctx, rd, projectID)
	return context.WithValue(ctx, requestDataKey{}, rd), rd
}

func setTraceData(ctx context.Context, md *requestData, projectID string) {
	traceID := md.requestID
	span := tracing.FromContext(ctx)
	if span != nil {
		sc := span.SpanContext()
		traceID = sc.TraceID.String()
		md.Entry.SpanID = sc.SpanID.String()
		md.Entry.TraceSampled = sc.IsSampled()
	}
	//Google expects a project prefix at the Entry level but not in the label
	md.commonLabels["appengine.googleapis.com/trace_id"] = traceID
	md.Entry.Trace = fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)
}

func requestDataFromContext(ctx context.Context) (md *requestData, ok bool) {
	md, ok = ctx.Value(requestDataKey{}).(*requestData)
	return
}
