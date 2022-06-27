package verrors

import (
	"google.golang.org/grpc/codes"
	"net/http"
)

// StatusCodeToGRPCError converts a http error into a grpc error
func StatusCodeToGRPCError(statusCode int) ErrorType {
	switch statusCode {
	case http.StatusBadRequest:
		return InvalidArgument
	case http.StatusUnauthorized:
		return Unauthenticated
	case http.StatusPaymentRequired:
		return FailedPrecondition
	case http.StatusForbidden:
		return PermissionDenied
	case http.StatusNotFound:
		return NotFound
	case http.StatusGone:
		return Gone
	case http.StatusConflict:
		return AlreadyExists
	case http.StatusPreconditionFailed:
		return FailedPrecondition
	case http.StatusTooManyRequests:
		return ResourceExhausted
	case http.StatusNotImplemented:
		return Unimplemented
	case http.StatusBadGateway:
		return BadGateway
	case http.StatusServiceUnavailable:
		return Unavailable
	default:
		return Internal
	}
}

// GRPCCodeToErrorType converts a grpc status code into the matching ErrorType or Unknown
func GRPCCodeToErrorType(statusCode codes.Code) ErrorType {
	switch statusCode {
	case codes.NotFound:
		return NotFound
	case codes.InvalidArgument:
		return InvalidArgument
	case codes.AlreadyExists:
		return AlreadyExists
	case codes.PermissionDenied:
		return PermissionDenied
	case codes.Unauthenticated:
		return Unauthenticated
	case codes.Unimplemented:
		return Unimplemented
	case codes.Unknown:
		return Unknown
	case codes.Internal:
		return Internal
	case codes.Unavailable:
		return Unavailable
	case codes.FailedPrecondition:
		return FailedPrecondition
	case codes.DeadlineExceeded:
		return DeadlineExceeded
	case codes.ResourceExhausted:
		return ResourceExhausted
	case codes.Aborted:
		return Aborted
	case codes.Canceled:
		return Canceled
	default:
		return Unknown
	}
}

//ErrorTypeToGRPCCode converts an error type into the matching grpc error code
func ErrorTypeToGRPCCode(errorType ErrorType) codes.Code {
	switch errorType {
	case NotFound:
		return codes.NotFound
	case InvalidArgument:
		return codes.InvalidArgument
	case AlreadyExists:
		return codes.AlreadyExists
	case PermissionDenied:
		return codes.PermissionDenied
	case Unauthenticated:
		return codes.Unauthenticated
	case Unimplemented:
		return codes.Unimplemented
	case Unknown:
		return codes.Unknown
	case Internal:
		return codes.Internal
	case Unavailable:
		return codes.Unavailable
	case FailedPrecondition:
		return codes.FailedPrecondition
	case DeadlineExceeded:
		return codes.DeadlineExceeded
	case ResourceExhausted:
		return codes.ResourceExhausted
	case Aborted:
		return codes.Aborted
	case Canceled:
		return codes.Canceled
	case BadGateway:
		return codes.Unavailable
	case Gone:
		return codes.FailedPrecondition
	default:
		return codes.Unknown
	}
}

// HttpStatusCodeToGRPCCode coverts a http status code into the matching grpc error code
func HttpStatusCodeToGRPCCode(statusCode int) codes.Code {
	return ErrorTypeToGRPCCode(StatusCodeToGRPCError(statusCode))
}
