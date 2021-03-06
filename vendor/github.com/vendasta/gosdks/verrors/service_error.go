package verrors

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServiceError is an error that can be translated to a GRPC-compliant error
type ServiceError struct {
	// This message will be logged by the error interceptor,
	// but will not be presented to users or shown in err.Error() calls.
	internalMsg string
	msg         string
	errType     ErrorType
}

// Error returns the message associated with this error
func (v ServiceError) Error() string {
	return v.msg
}

// ErrorType returns the ErrorType associated with this error
func (v ServiceError) ErrorType() ErrorType {
	return v.errType
}

// GRPCError returns an error in a format such that it can be consumed by GRPC
func (v ServiceError) GRPCError() error {
	grpcCode := ErrorTypeToGRPCCode(v.errType)
	if grpcCode == codes.Unknown {
		return status.Errorf(codes.Unknown, "Unknown server error.")
	}
	return status.Errorf(grpcCode, v.msg)
}

// HTTPCode returns the corresponding http status code for a given error
func (v ServiceError) HTTPCode() int {
	switch v.errType {
	case NotFound:
		return http.StatusNotFound
	case InvalidArgument:
		return http.StatusBadRequest
	case AlreadyExists:
		return http.StatusConflict
	case PermissionDenied:
		return http.StatusForbidden
	case Unauthenticated:
		return http.StatusUnauthorized
	case Unimplemented:
		return http.StatusNotImplemented
	case FailedPrecondition:
		return http.StatusPreconditionFailed
	case DeadlineExceeded:
		return http.StatusRequestTimeout
	case ResourceExhausted:
		return http.StatusTooManyRequests
	case Unavailable:
		return http.StatusServiceUnavailable
	case Aborted:
		return http.StatusConflict
	case Canceled:
		return StatusClientClosedRequest
	case BadGateway:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// Add internal notes to an error which will be logged, but never shown to the end user.
func (v ServiceError) WithInternalMessage(internalMsg string) ServiceError {
	if v.internalMsg != "" {
		v.internalMsg = fmt.Sprintf("%s. %s", v.internalMsg, internalMsg)
		return v
	}
	v.internalMsg = internalMsg
	return v
}

// Get internal notes from an error, these should not be shown to end users.
func (v ServiceError) GetInternalMessage() string {
	return v.internalMsg
}

// New returns a ServiceError
func New(errorType ErrorType, format string, a ...interface{}) ServiceError {
	return ServiceError{msg: fmt.Sprintf(format, a...), errType: errorType}
}

// FromErrorWithContext is the same as FromError, except that if the ctx has
// been canceled or if its deadline has been exceeded, it returns Canceled or
// DeadlineExceeded, respectively.
func FromErrorWithContext(ctx context.Context, err error) ServiceError {
	if ctx.Err() == context.Canceled {
		err = New(Canceled, "Request was canceled")
	}
	if ctx.Err() == context.DeadlineExceeded {
		err = New(DeadlineExceeded, "Deadline exceeded")
	}
	return FromError(err)
}

// FromError given an error tries to return a proper ServiceError.
func FromError(err error) ServiceError {
	if err == context.Canceled {
		err = New(Canceled, err.Error())
	}
	if err == context.DeadlineExceeded {
		err = New(DeadlineExceeded, err.Error())
	}

	statusErr, ok := status.FromError(err)
	if ok {
		return New(GRPCCodeToErrorType(statusErr.Code()), statusErr.Message())
	}

	serviceError, ok := err.(ServiceError)
	if ok {
		return serviceError
	}
	return New(Unknown, "Unknown server error.")
}

// WrapError adds additional messages to an error without affecting the ErrorType of Service Errors or GRPC Errors.
// Do not use WrapError if there is a need for metadata held by the originating error that may be needed later.
func WrapError(err error, format string, args ...interface{}) ServiceError {
	newMsg := fmt.Sprintf(format, args...)
	errMsg := fmt.Sprintf("%s: %s", newMsg, err.Error())
	svcErr := FromError(err)
	return New(svcErr.ErrorType(), errMsg).WithInternalMessage(svcErr.internalMsg)
}

// IsError returns true/false if the given err matches the errorType type.
// If err == nil, this function always returns false.
func IsError(errorType ErrorType, err error) bool {
	if err == nil {
		return false
	}
	statusErr, ok := status.FromError(err)
	if ok {
		return GRPCCodeToErrorType(statusErr.Code()) == errorType
	}

	serviceError, ok := err.(ServiceError)
	if ok {
		return serviceError.errType == errorType
	}
	return false
}

// ToGrpcError calculates the correct GRPC error code for a ServiceError or existing GRPC error and returns it
// All errors that are not GRPC errors or ServiceErrors will be interpreted as Unknown errors
func ToGrpcError(err error) error {
	// if this is already a GRPC error, pass through
	grpcErr, ok := status.FromError(err)
	if ok {
		return grpcErr.Err()
	}
	// otherwise map to ServiceError
	return FromError(err).GRPCError()
}

// ToGrpcErrorWithContext examines the errors inside the given context and translates them to a GRPC error.
func ToGrpcErrorWithContext(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	if ctx.Err() == context.Canceled {
		return status.Errorf(codes.Canceled, "context canceled %s", err.Error())
	}
	if ctx.Err() == context.DeadlineExceeded {
		return status.Errorf(codes.DeadlineExceeded, "deadline exceeded %s", err.Error())
	}

	return ToGrpcError(err)
}
