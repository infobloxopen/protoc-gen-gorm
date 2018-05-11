package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// Header returns first value for a given key if it exists in gRPC metadata
// from incoming or outcoming context, otherwise returns (nil, false)
//
// Calls HeaderN(ctx, key, 1)
//
// Provided key is converted to lowercase (see grpc/metadata.New).
// If key is not found the prefix "grpcgateway-" is added to the key and
// key is being searched once again.
func Header(ctx context.Context, key string) (string, bool) {
	if l, ok := HeaderN(ctx, key, 1); ok {
		return l[0], ok
	}
	return "", false
}

// HeaderN returns first n values for a given key if it exists in gRPC metadata
// from incoming or outcoming context, otherwise returns (nil, false)
//
// If n < 0 all values for a given key will be returned
// If n > 0 at least n values will be returned, or (nil, false)
// If n == 0 result is (nil, false)
//
// Provided key is converted to lowercase (see grpc/metadata.New).
// If key is not found the prefix "grpcgateway-" is added to the key and
// key is being searched once again.
func HeaderN(ctx context.Context, key string, n int) (val []string, found bool) {
	if n == 0 {
		return
	}

	if smd, ok := runtime.ServerMetadataFromContext(ctx); ok {
		ctx = metadata.NewIncomingContext(ctx, smd.HeaderMD)
	}

	imd, iok := metadata.FromIncomingContext(ctx)
	omd, ook := metadata.FromOutgoingContext(ctx)

	md := metadata.Join(imd, omd)

	if !iok && !ook {
		return nil, false
	}

	key = strings.ToLower(key)
	if v, ok := md[key]; ok {
		val = append(val, v...)
		found = true
	}
	// If md contains 'key' and 'runtime.MetadataPrefix + key'
	// collect them all
	key = runtime.MetadataPrefix + key
	if v, ok := md[key]; ok {
		val = append(val, v...)
		found = true
	}

	switch {
	case !found:
		return
	case n < 0 || len(val) == n:
		return
	case len(val) < n:
		return nil, false
	default:
		return val[:n], found
	}
}

// PrefixOutgoingHeaderMatcher prefixes outgoing gRPC metadata with
// runtime.MetadataHeaderPrefix ("Grpc-Metadata-").
// It behaves like the default gRPC-Gateway outgoing header matcher
// (if none is provided as an option).
func PrefixOutgoingHeaderMatcher(key string) (string, bool) {
	return fmt.Sprintf("%s%s", runtime.MetadataHeaderPrefix, key), true
}

func handleForwardResponseServerMetadata(matcher runtime.HeaderMatcherFunc, w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.HeaderMD {
		if h, ok := matcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}
