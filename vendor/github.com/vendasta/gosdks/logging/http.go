package logging

import (
	"net/http"
	"time"
)

func newLoggedResponse(w http.ResponseWriter) *loggedResponse {
	return &loggedResponse{w, 200, 0}
}

type loggedResponse struct {
	http.ResponseWriter
	status int
	length int
}

func (l *loggedResponse) WriteHeader(status int) {
	l.status = status
	l.ResponseWriter.WriteHeader(status)
}

func (l *loggedResponse) Write(b []byte) (n int, err error) {
	n, err = l.ResponseWriter.Write(b)
	l.length += n
	return
}

// HTTPMiddleware provides logging/tracing for incoming http requests.
func HTTPMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		ctx = NewTaggedContext(ctx)

		// Add some useful context to logs
		if origin := request.Header.Get("Origin"); origin != "" {
			Tag(ctx, "origin", origin)
		}
		if referer := request.Header.Get("Referer"); referer != "" {
			Tag(ctx, "referer", referer)
		}
		if host := request.Header.Get("Host"); host != "" {
			Tag(ctx, "host", host)
		}

		ctx, requestData := newRequest(ctx, GetLogger().RequestID())
		request = request.WithContext(ctx)

		response := newLoggedResponse(w)

		start := time.Now()
		h.ServeHTTP(response, request)
		end := time.Now()
		latency := end.Sub(start)

		ctx = applyBundlingMetadata(ctx, requestData, request, response, latency)
		logRequest(ctx, requestData, Request)
	})
}
