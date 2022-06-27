package logging

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vendasta/gosdks/verrors"
)

const fallbackPath = "/withBundling/default/path"

// WithBundling runs the given function with a context set up for log bundling to the given stream
// The status code for the log bundle for non-nil errors is computed as verrors.FromError(err).HttpCode()
// WithBundling() runs the given function synchronously and returns any error from it.
//
// path is used as the URL in the bundled request log, set it to a unique identifier (even for non-request-based code)
// stream is the log-stream to send logs to; e.g. Pubsub, Request, etc.
// Example usage for bundling logging for an executive report cron job:
//
// logging.WithBundling(ctx, "/cron/executive-report/weekly", Background, func(ctx context.Context) error {
//   // do your work here, using the ctx which was passed in
//   // return a verrors.ServiceError if things fail
// })
func WithBundling(ctx context.Context, path string, stream Stream, work func(context.Context) error) error {
	ctx, requestData := newRequest(ctx, GetLogger().RequestID())
	urlObj, err := url.Parse(path)
	if err != nil {
		urlObj, _ = url.Parse(fallbackPath)
	}

	start := time.Now()
	workError := work(ctx)
	end := time.Now()

	latency := end.Sub(start)
	resp := &loggedResponse{status: 200}
	req := &http.Request{URL: urlObj, Method: strings.ToUpper(string(stream))}

	if workError != nil {
		resp.status = verrors.FromError(workError).HTTPCode()
	}

	ctx = applyBundlingMetadata(ctx, requestData, req, resp, latency)
	logRequest(ctx, requestData, stream)
	return workError
}
