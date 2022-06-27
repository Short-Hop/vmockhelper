package logging

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/vendasta/gosdks/tracing"
	"google.golang.org/grpc/metadata"
)

// applyRequestMetadata sets request metadata on a requestData object
func applyRequestMetadata(rd *requestData, request *http.Request, response *loggedResponse, latency time.Duration) {
	if request != nil {
		rd.HTTPRequest.Request = request
		rd.HTTPRequest.RemoteIP = request.RemoteAddr
	}

	if response != nil {
		rd.HTTPRequest.Status = int(response.status)
		rd.HTTPRequest.ResponseSize = int64(response.length)
	}

	if latency != 0 {
		rd.HTTPRequest.Latency = latency
	}

	rd.HTTPRequest.LocalIP = "127.0.0.1"
}
func applyBundlingMetadata(ctx context.Context, rd *requestData,
	request *http.Request, response *loggedResponse, latency time.Duration) context.Context {
	applyRequestMetadata(rd, request, response, latency)
	md, _ := metadata.FromOutgoingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}

func calculateTraceID(ctx context.Context, config *config) string {
	projectID := "unknown"
	if config != nil && config.ProjectID != "" {
		projectID = config.ProjectID
	}
	traceID := tracing.FromContext(ctx).SpanContext().TraceID.String()
	if traceID != missingTraceID {
		return fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)
	}

	rd, ok := requestDataFromContext(ctx)
	if ok {
		if rd.Trace != "" {
			return rd.Trace
		}
		return rd.requestID
	}
	return ""
}
