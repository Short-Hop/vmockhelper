package logging

import (
	"context"
	"github.com/vendasta/gosdks/verrors"
)

// ErrorWithTrace returns a ServiceError and logs its stack trace
func ErrorWithTrace(ctx context.Context, errorType verrors.ErrorType, format string, a ...interface{}) verrors.ServiceError {
	StackTrace(ctx, format, a...)
	return verrors.New(errorType, format, a...)
}
