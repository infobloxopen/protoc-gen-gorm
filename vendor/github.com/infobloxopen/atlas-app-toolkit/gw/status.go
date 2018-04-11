package gw

import (
	"context"
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

const (
	// These custom codes defined here to conform REST API Syntax
	// It is supposed that you do not send them over the wire as part of gRPC Status,
	// because they will be treated as Unknown by gRPC library.
	// You should use them to send successfull status of your RPC method
	// using SetStatus function from this package.
	Created codes.Code = 10000 + iota // 10000 is an offset from standard codes
	Updated
	Deleted
	LongRunning
	PartialContent
)

// RestStatus represents a response status in accordance with REST API Syntax.
// See: https://github.com/infobloxopen/atlas-app-toolkit#responses
type RestStatus struct {
	HTTPStatus int `json:"status,omitempty"`
	// Code is a string representation of an error code
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`

	// Pagination response parameters
	PageToken string `json:"_page_token,omitempty"`
	Offset    string `json:"_offset,omitempty"`
	Size      string `json:"_size,omitempty"`
}

// SetStatus sets gRPC status as gRPC metadata
// Status.Code will be set with metadata key `grpcgateway-status-code` and
// with value as string name of the code.
// Status.Message will be set with metadata key `grpcgateway-status-message`
// and with corresponding value.
func SetStatus(ctx context.Context, st *status.Status) error {
	if st == nil {
		return nil
	}

	md := metadata.Pairs(
		runtime.MetadataPrefix+"status-code", CodeName(st.Code()),
		runtime.MetadataPrefix+"status-message", st.Message(),
	)
	return grpc.SetHeader(ctx, md)
}

// SetCreated is a shortcut for SetStatus(ctx, status.New(Created, msg))
func SetCreated(ctx context.Context, msg string) error {
	return SetStatus(ctx, status.New(Created, msg))
}

// SetUpdated is a shortcut for SetStatus(ctx, status.New(Updated, msg))
func SetUpdated(ctx context.Context, msg string) error {
	return SetStatus(ctx, status.New(Updated, msg))
}

// SetDeleted is a shortcut for SetStatus(ctx, status.New(Deleted, msg))
func SetDeleted(ctx context.Context, msg string) error {
	return SetStatus(ctx, status.New(Deleted, msg))
}

// SetRunning is a shortcut for SetStatus(ctx, status.New(LongRunning, url))
func SetRunning(ctx context.Context, message, resource string) error {
	grpc.SetHeader(ctx, metadata.Pairs("Location", resource))
	return SetStatus(ctx, status.New(LongRunning, message))
}

// Status returns REST representation of gRPC status.
// If status.Status is not nil it will be converted in accrodance with REST
// API Syntax otherwise context will be used to extract
// `grpcgatewau-status-code` and `grpcgateway-status-message` from
// gRPC metadata.
// If `grpcgatewau-status-code` is not set it is assumed that it is OK.
func Status(ctx context.Context, st *status.Status) *RestStatus {
	var rst RestStatus

	if st != nil {
		rst.Code = CodeName(st.Code())
		rst.HTTPStatus = HTTPStatusFromCode(st.Code())
		rst.Message = st.Message()

		return &rst
	}

	if sc, ok := Header(ctx, "status-code"); ok {
		rst.Code = sc
	} else {
		rst.Code = CodeName(codes.OK)
	}
	if sm, ok := Header(ctx, "status-message"); ok {
		rst.Message = sm
	}
	rst.HTTPStatus = HTTPStatusFromCode(Code(rst.Code))

	// PageInfo
	if pt, ok := Header(ctx, pageInfoPageTokenMetaKey); ok {
		rst.PageToken = pt
	}
	if o, ok := Header(ctx, pageInfoOffsetMetaKey); ok {
		rst.Offset = o
	}
	if s, ok := Header(ctx, pageInfoSizeMetaKey); ok {
		rst.Size = s
	}

	return &rst
}

// CodeName returns stringname of gRPC code, function handles as standard
// codes from "google.golang.org/grpc/codes" as well as custom ones defined
// in this package.
// The codes.Unimplemented is named "NOT_IMPLEMENTED" in accordance with
// REST API Syntax Specification.
func CodeName(c codes.Code) string {
	switch c {
	case codes.Unimplemented:
		return "NOT_IMPLEMENTED"
	case Created:
		return "CREATED"
	case Updated:
		return "UPDATED"
	case Deleted:
		return "DELETED"
	case LongRunning:
		return "LONG_RUNNING_OP"
	case PartialContent:
		return "PARTIAL_CONTENT"
	default:
		var cname string
		if cn, ok := code.Code_name[int32(c)]; !ok {
			cname = code.Code_UNKNOWN.String()
		} else {
			cname = cn
		}
		return cname
	}
}

// Code returns an instance of gRPC code by its string name.
// The `cname` must be in upper case and one of the code names
// defined in REST API Syntax.
// If code name is invalid or unknow the codes.Unknown will be returned.
func Code(cname string) codes.Code {
	switch cname {
	case "NOT_IMPLEMENTED":
		return codes.Unimplemented
	case "CREATED":
		return Created
	case "UPDATED":
		return Updated
	case "DELETED":
		return Deleted
	case "LONG_RUNNING_OP":
		return LongRunning
	case "PARTIAL_CONTENT":
		return PartialContent
	default:
		var c codes.Code
		if cc, ok := code.Code_value[cname]; !ok {
			c = codes.Unknown
		} else {
			c = codes.Code(cc)
		}
		return c
	}
}

// HTTPStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case Created:
		return http.StatusCreated
	case Updated:
		return http.StatusCreated
	case Deleted:
		return http.StatusNoContent
	case LongRunning:
		return http.StatusAccepted
	case PartialContent:
		return http.StatusPartialContent
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499 // (gRPC-Gateway - http.StatusRequestTimeout = 408)
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout // = 504 (gRPC-Gateway - http.StatusRequestTimeout = 408)
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests // = 429 (gRPC-Gateway - http.StatusForbidden = 403)
	case codes.FailedPrecondition:
		return http.StatusBadRequest // = 400 (gRPC-Gateway - http.StatusPreconditionFailed = 412)
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}

	grpclog.Printf("Unknown gRPC error code: %v", code)
	return http.StatusInternalServerError
}
