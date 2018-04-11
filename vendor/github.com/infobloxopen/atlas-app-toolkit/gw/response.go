package gw

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

type (
	// ForwardResponseMessageFunc forwards gRPC response to HTTP client inaccordance with REST API Syntax
	ForwardResponseMessageFunc func(context.Context, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, proto.Message, ...func(context.Context, http.ResponseWriter, proto.Message) error)
	// ForwardResponseStreamFunc forwards gRPC stream response to HTTP client inaccordance with REST API Syntax
	ForwardResponseStreamFunc func(context.Context, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, func() (proto.Message, error), ...func(context.Context, http.ResponseWriter, proto.Message) error)
)

// ResponseForwarder implements ForwardResponseMessageFunc in method ForwardMessage
// and ForwardResponseStreamFunc in method ForwardStream
// in accordance with REST API Syntax Specification.
// See: https://github.com/infobloxopen/atlas-app-toolkit#responses
// for format of JSON response.
type ResponseForwarder struct {
	OutgoingHeaderMatcher runtime.HeaderMatcherFunc
	MessageErrHandler     runtime.ProtoErrorHandlerFunc
	StreamErrHandler      ProtoStreamErrorHandlerFunc
}

var (
	// ForwardResponseMessage is default implementation of ForwardResponseMessageFunc
	ForwardResponseMessage = NewForwardResponseMessage(PrefixOutgoingHeaderMatcher, ProtoMessageErrorHandler, ProtoStreamErrorHandler)
	// ForwardResponseStream is default implementation of ForwardResponseStreamFunc
	ForwardResponseStream = NewForwardResponseStream(PrefixOutgoingHeaderMatcher, ProtoMessageErrorHandler, ProtoStreamErrorHandler)
)

// NewForwardResponseMessage returns ForwardResponseMessageFunc
func NewForwardResponseMessage(out runtime.HeaderMatcherFunc, meh runtime.ProtoErrorHandlerFunc, seh ProtoStreamErrorHandlerFunc) ForwardResponseMessageFunc {
	fw := &ResponseForwarder{out, meh, seh}
	return fw.ForwardMessage
}

// NewForwardResponseStream returns ForwardResponseStreamFunc
func NewForwardResponseStream(out runtime.HeaderMatcherFunc, meh runtime.ProtoErrorHandlerFunc, seh ProtoStreamErrorHandlerFunc) ForwardResponseStreamFunc {
	fw := &ResponseForwarder{out, meh, seh}
	return fw.ForwardStream
}

// ForwardMessage implements runtime.ForwardResponseMessageFunc
func (fw *ResponseForwarder) ForwardMessage(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, resp proto.Message, opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Printf("forward response message: failed to extract ServerMetadata from context")
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, fmt.Errorf("forward response message: internal error"))
	}

	handleForwardResponseServerMetadata(fw.OutgoingHeaderMatcher, rw, md)
	handleForwardResponseTrailerHeader(rw, md)

	rw.Header().Set("Content-Type", marshaler.ContentType())

	if err := handleForwardResponseOptions(ctx, rw, resp, opts); err != nil {
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
		return
	}

	// here we start doing a bit strange things
	// 1. marshal response into bytes
	// 2. unmarshal bytes into dynamic map[string]interface{}
	// 3. add our custom metadata into dynamic map
	// 4. marshal dynamic map into bytes again :\
	// all that steps are needed because of this requirements:
	// -- To allow compatibility with existing systems,
	// -- the results tag name can be changed to a service-defined tag.
	// -- In this way the success data becomes just a tag added to an existing structure.
	data, err := marshaler.Marshal(resp)
	if err != nil {
		grpclog.Printf("forward response: failed to marshal response: %v", err)
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
	}

	var dynmap map[string]interface{}
	if err := marshaler.Unmarshal(data, &dynmap); err != nil {
		grpclog.Printf("forward response: failed to unmarshal response: %v", err)
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
	}

	retainFields(ctx, req, dynmap)

	// Here we set "Location" header which contains a url to a long running task
	// Using it we can retrieve its status
	rst := Status(ctx, nil)
	if rst.Code == CodeName(LongRunning) {
		location, exists := Header(ctx, "Location")

		if !exists || location == "" {
			err := fmt.Errorf("Header Location should be set for long running operation")
			grpclog.Printf("forward response: %v", err)
			fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
		}
		rw.Header().Add("Location", location)
	}
	// this is the edge case, if user sends response that has field 'success'
	// let him see his response object instead of our status
	if _, ok := dynmap["success"]; !ok {
		dynmap["success"] = rst
	}

	data, err = marshaler.Marshal(dynmap)
	if err != nil {
		grpclog.Printf("forward response: failed to marshal response: %v", err)
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
	}

	rw.WriteHeader(rst.HTTPStatus)

	if _, err = rw.Write(data); err != nil {
		grpclog.Printf("forward response: failed to write response: %v", err)
	}

	handleForwardResponseTrailer(rw, md)
}

// ForwardStream implements runtime.ForwardResponseStreamFunc.
// RestStatus comes first in the chuncked result.
func (fw *ResponseForwarder) ForwardStream(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, recv func() (proto.Message, error), opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		grpclog.Printf("forward response stream: flush not supported in %T", rw)
		fw.StreamErrHandler(ctx, false, mux, marshaler, rw, req, fmt.Errorf("forward response message: internal error"))
		return
	}

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Printf("forward response stream: failed to extract ServerMetadata from context")
		fw.StreamErrHandler(ctx, false, mux, marshaler, rw, req, fmt.Errorf("forward response message: internal error"))
		return
	}
	handleForwardResponseServerMetadata(fw.OutgoingHeaderMatcher, rw, md)

	rw.Header().Set("Transfer-Encoding", "chunked")
	rw.Header().Set("Content-Type", marshaler.ContentType())

	if err := handleForwardResponseOptions(ctx, rw, nil, opts); err != nil {
		fw.StreamErrHandler(ctx, false, mux, marshaler, rw, req, err)
		return
	}

	rst := Status(ctx, nil)
	// if user did not set status explicitly
	if rst.Code == "" || rst.Code == CodeName(codes.OK) {
		rst.Code = CodeName(PartialContent)
	}
	if rst.HTTPStatus == http.StatusOK {
		rst.HTTPStatus = HTTPStatusFromCode(PartialContent)
	}
	v := map[string]interface{}{"success": rst}

	rw.WriteHeader(rst.HTTPStatus)

	data, err := marshaler.Marshal(v)
	if err != nil {
		fw.StreamErrHandler(ctx, true, mux, marshaler, rw, req, err)
		return
	}

	if _, err := rw.Write(data); err != nil {
		grpclog.Printf("forward response stream: failed to write status object: %s", err)
		return
	}

	var delimiter []byte
	if d, ok := marshaler.(runtime.Delimited); ok {
		delimiter = d.Delimiter()
	} else {
		delimiter = []byte("\n")
	}

	for {
		resp, err := recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			fw.StreamErrHandler(ctx, true, mux, marshaler, rw, req, err)
			return
		}
		if err := handleForwardResponseOptions(ctx, rw, resp, opts); err != nil {
			fw.StreamErrHandler(ctx, true, mux, marshaler, rw, req, err)
			return
		}

		data, err := marshaler.Marshal(resp)
		if err != nil {
			fw.StreamErrHandler(ctx, true, mux, marshaler, rw, req, err)
			return
		}

		if _, err := rw.Write(data); err != nil {
			grpclog.Printf("forward response stream: failed to write response object: %s", err)
			return
		}

		if _, err = rw.Write(delimiter); err != nil {
			grpclog.Printf("forward response stream: failed to send delimiter chunk: %v", err)
			return
		}
		flusher.Flush()
	}
}

func handleForwardResponseOptions(ctx context.Context, rw http.ResponseWriter, resp proto.Message, opts []func(context.Context, http.ResponseWriter, proto.Message) error) error {
	if len(opts) == 0 {
		return nil
	}
	for _, opt := range opts {
		if err := opt(ctx, rw, resp); err != nil {
			grpclog.Printf("error handling ForwardResponseOptions: %v", err)
			return err
		}
	}
	return nil
}
