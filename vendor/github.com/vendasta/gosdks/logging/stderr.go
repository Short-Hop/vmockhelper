package logging

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"cloud.google.com/go/logging"
)

const (
	colorRed = iota + 30
	colorGreen
	colorYellow
	colorMagenta
	colorCyan
)

type color int

var (
	colors = []string{
		logging.Debug:     colorSeq(colorMagenta),
		logging.Info:      colorSeq(colorRed),
		logging.Warning:   colorSeq(colorYellow),
		logging.Error:     colorSeq(colorGreen),
		logging.Critical:  colorSeq(colorCyan),
		logging.Alert:     colorSeq(colorCyan),
		logging.Emergency: colorSeq(colorCyan),
	}
)

func colorSeq(color color) string {
	return fmt.Sprintf("\033[%dm", int(color))
}

type stderrLogger struct {
	config *config
}

// Compile time interface check
var _ Logger = &stderrLogger{}

func (l *stderrLogger) request(ctx context.Context, r *requestData, stream Stream) {
	path := ""
	if r != nil && r.HTTPRequest != nil && r.HTTPRequest.Request != nil && r.HTTPRequest.Request.URL != nil {
		path = r.HTTPRequest.Request.URL.Path
	}
	status := 0
	if r != nil && r.HTTPRequest != nil {
		status = r.HTTPRequest.Status
	}
	l.Log(ctx, logging.Debug, "Served gRPC request for handler %s with code %d. Trace: %v", path, status, r.Trace)
}
func (l *stderrLogger) Debugf(ctx context.Context, f string, a ...interface{}) {
	l.Log(ctx, logging.Debug, f, a...)
}

func (l *stderrLogger) Infof(ctx context.Context, f string, a ...interface{}) {
	l.Log(ctx, logging.Info, f, a...)
}

func (l *stderrLogger) Warningf(ctx context.Context, f string, a ...interface{}) {
	l.Log(ctx, logging.Warning, f, a...)
}

func (l *stderrLogger) Errorf(ctx context.Context, f string, a ...interface{}) {
	l.Log(ctx, logging.Error, f, a...)
}

func (l *stderrLogger) Criticalf(ctx context.Context, f string, a ...interface{}) {
	l.Log(ctx, logging.Critical, f, a...)
}

func (l *stderrLogger) Alertf(ctx context.Context, f string, a ...interface{}) {
	l.Log(ctx, logging.Alert, f, a...)
}

func (l *stderrLogger) Emergencyf(ctx context.Context, f string, a ...interface{}) {
	l.Log(ctx, logging.Emergency, f, a...)
}

func (l *stderrLogger) StackTrace(ctx context.Context, f string, a ...interface{}) {
	message := fmt.Sprintf(f, a...)
	stackTrace := debug.Stack()
	messageWithStack := fmt.Sprintf("%s:\n%s", message, string(stackTrace))
	l.Debugf(ctx, messageWithStack)
}

func (m *stderrLogger) Tag(ctx context.Context, key, value string) Logger { return m }

func (l *stderrLogger) Log(ctx context.Context, level logging.Severity, f string, a ...interface{}) {

	// If the severity of the log is less then the log inclusion configured, don't log
	if level < logging.Severity(l.config.loggingInclusionLevel) {
		return
	}

	col := colors[level]
	var file string
	var line int
	if level >= logging.Severity(l.config.filenameLoggingLevel) {
		_, f, l, ok := runtime.Caller(3)
		if ok {
			filePieces := strings.Split(f, "/")
			file = strings.Join(filePieces[len(filePieces)-2:], "/")
			line = l
		}
	}
	if !strings.HasSuffix(f, "\n") {
		f = f + "\n"
	}
	var prefix string
	if line > 0 {
		prefix = fmt.Sprintf("%s%-9s%50s:%-4d\033[0m", col, level.String(), file, line)
	} else {
		prefix = fmt.Sprintf("%s%-9s%50s:    \033[0m", col, level.String(), file)
	}
	msg := prefix + " " + f
	if len(a) > 0 {
		fmt.Fprintf(os.Stderr, msg, a...)
	} else {
		os.Stderr.WriteString(msg)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (l *stderrLogger) RequestID() string {
	b := make([]rune, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (l *stderrLogger) ActiveRequestId(ctx context.Context) string {
	rd, ok := requestDataFromContext(ctx)
	if !ok {
		return ""
	}
	return rd.requestID
}

func (l *stderrLogger) ActiveTraceId(ctx context.Context) string {
	return calculateTraceID(ctx, l.config)
}

func newStdErrLogger(config *config) (*stderrLogger, error) {
	return &stderrLogger{config: config}, nil
}
