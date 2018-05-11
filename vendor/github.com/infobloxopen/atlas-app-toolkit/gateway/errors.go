package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// ProtoStreamErrorHandlerFunc handles the error as a gRPC error generated via status package and replies to the testRequest.
// Addition bool argument indicates whether method (http.ResponseWriter.WriteHeader) was called or not.
type ProtoStreamErrorHandlerFunc func(context.Context, bool, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, error)

// RestError represents an error in accordance with REST API Syntax Specification.
// See: https://github.com/infobloxopen/atlas-app-toolkit#errors
type RestError struct {
	Status  *RestStatus   `json:"error,omitempty"`
	Details []interface{} `json:"details,omitempty"`
}

var (
	// ProtoMessageErrorHandler uses PrefixOutgoingHeaderMatcher.
	// To use ProtoErrorHandler with custom outgoing header matcher call NewProtoMessageErrorHandler.
	ProtoMessageErrorHandler = NewProtoMessageErrorHandler(PrefixOutgoingHeaderMatcher)
	// ProtoStreamErrorHandler uses PrefixOutgoingHeaderMatcher.
	// To use ProtoErrorHandler with custom outgoing header matcher call NewProtoStreamErrorHandler.
	ProtoStreamErrorHandler = NewProtoStreamErrorHandler(PrefixOutgoingHeaderMatcher)
)

// NewProtoMessageErrorHandler returns runtime.ProtoErrorHandlerFunc
func NewProtoMessageErrorHandler(out runtime.HeaderMatcherFunc) runtime.ProtoErrorHandlerFunc {
	h := &ProtoErrorHandler{out}
	return h.MessageHandler
}

// NewProtoStreamErrorHandler returns ProtoStreamErrorHandlerFunc
func NewProtoStreamErrorHandler(out runtime.HeaderMatcherFunc) ProtoStreamErrorHandlerFunc {
	h := &ProtoErrorHandler{out}
	return h.StreamHandler
}

// ProtoErrorHandler implements runtime.ProtoErrorHandlerFunc in method MessageHandler
// and ProtoStreamErrorHandlerFunc in method StreamHandler
// in accordance with REST API Syntax Specification.
// See RestError for the JSON format of an error
type ProtoErrorHandler struct {
	OutgoingHeaderMatcher runtime.HeaderMatcherFunc
}

// MessageHandler implements runtime.ProtoErrorHandlerFunc
// in accordance with REST API Syntax Specification.
// See RestError for the JSON format of an error
func (h *ProtoErrorHandler) MessageHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, err error) {

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Printf("error handler: failed to extract ServerMetadata from context")
	}

	handleForwardResponseServerMetadata(h.OutgoingHeaderMatcher, rw, md)
	handleForwardResponseTrailerHeader(rw, md)

	h.writeError(ctx, false, marshaler, rw, err)

	handleForwardResponseTrailer(rw, md)
}

// StreamHandler implements ProtoStreamErrorHandlerFunc
// in accordance with REST API Syntax Specification.
// See RestError for the JSON format of an error
func (h *ProtoErrorHandler) StreamHandler(ctx context.Context, headerWritten bool, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, err error) {
	h.writeError(ctx, headerWritten, marshaler, rw, err)
}

func (h *ProtoErrorHandler) writeError(ctx context.Context, headerWritten bool, marshaler runtime.Marshaler, rw http.ResponseWriter, err error) {
	const fallback = `{"code":"INTERNAL","status":500,"message":"%s"}`

	st, ok := status.FromError(err)
	if !ok {
		st = status.New(codes.Unknown, err.Error())
	}

	restErr := &RestError{
		Status:  Status(ctx, st),
		Details: st.Details(),
	}

	if !headerWritten {
		rw.Header().Del("Trailer")
		rw.Header().Set("Content-Type", marshaler.ContentType())
		rw.WriteHeader(restErr.Status.HTTPStatus)
	}

	buf, merr := marshaler.Marshal(restErr)
	if merr != nil {
		grpclog.Printf("error handler: failed to marshal error message %q: %v", restErr, merr)
		rw.WriteHeader(http.StatusInternalServerError)

		if _, err := io.WriteString(rw, fmt.Sprintf(fallback, merr)); err != nil {
			grpclog.Printf("error handler: failed to write response: %v", err)
		}
		return
	}

	if _, err := rw.Write(buf); err != nil {
		grpclog.Printf("error handler: failed to write response: %v", err)
	}
}
