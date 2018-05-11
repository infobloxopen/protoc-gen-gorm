package gateway

import (
	"context"
	"net/http"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

const (
	filterQueryKey           = "_filter"
	filterMetaKey            = "operator-filter"
	sortQueryKey             = "_order_by"
	sortMetaKey              = "operator-sort"
	fieldsQueryKey           = "_fields"
	fieldsMetaKey            = "operator-fields"
	limitQueryKey            = "_limit"
	limitMetaKey             = "operator-limit"
	offsetQueryKey           = "_offset"
	offsetMetaKey            = "operator-offset"
	pageTokenQueryKey        = "_page_token"
	pageTokenMetaKey         = "operator-page-token"
	pageInfoSizeMetaKey      = "status-page-info-size"
	pageInfoOffsetMetaKey    = "status-page-info-offset"
	pageInfoPageTokenMetaKey = "status-page-info-page_token"
)

// MetadataAnnotator is a function for passing metadata to a gRPC context
// It must be mainly used as ServeMuxOption for gRPC Gateway 'ServeMux'
// See: 'WithMetadata' option.
//
// MetadataAnnotator extracts values of query operators from incoming
// HTTP testRequest accroding to REST API Syntax.
// E.g:
// - _order_by="name asc,age desc"
// - _fields="name,age"
// - _filter="name == 'John'"
// - _limit=1000
// - _offset=1001
// - _page_token=QWxhZGRpbjpvcGVuIHNlc2FtZQ
func MetadataAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	vals := req.URL.Query()
	mdmap := make(map[string]string)

	if v := vals.Get(sortQueryKey); v != "" {
		mdmap[runtime.MetadataPrefix+sortMetaKey] = v
	}
	if v := vals.Get(fieldsQueryKey); v != "" {
		mdmap[runtime.MetadataPrefix+fieldsMetaKey] = v
	}

	if v := vals.Get(filterQueryKey); v != "" {
		mdmap[runtime.MetadataPrefix+filterMetaKey] = v
	}

	if v := vals.Get(offsetQueryKey); v != "" {
		mdmap[runtime.MetadataPrefix+offsetMetaKey] = v
	}

	if v := vals.Get(limitQueryKey); v != "" {
		mdmap[runtime.MetadataPrefix+limitMetaKey] = v
	}

	if v := vals.Get(pageTokenQueryKey); v != "" {
		mdmap[runtime.MetadataPrefix+pageTokenMetaKey] = v
	}

	return metadata.New(mdmap)
}

// Sorting extracts sort parameters from incoming gRPC context.
// If sorting collection operator has not been specified in query string of
// incoming HTTP testRequest function returns (nil, nil).
// If provided sorting parameters are invalid function returns
// `status.Error(codes.InvalidArgument, parser_error)`
// See: `query.ParseSorting` for details.
func Sorting(ctx context.Context) (*query.Sorting, error) {
	raw, ok := Header(ctx, sortMetaKey)
	if !ok {
		return nil, nil
	}

	s, err := query.ParseSorting(raw)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return s, nil
}

// Filtering extracts filter parameters from incoming gRPC context.
// If filtering collection operator has not been specified in query string of
// incoming HTTP testRequest function returns (nil, nil).
// If provided filtering parameters are invalid function returns
// `status.Error(codes.InvalidArgument, parser_error)`
// See: `query.ParseFiltering` for details.
func Filtering(ctx context.Context) (*query.Filtering, error) {
	raw, ok := Header(ctx, filterMetaKey)
	if !ok {
		return nil, nil
	}

	f, err := query.ParseFiltering(raw)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return f, nil
}

// Pagination extracts pagination parameters from incoming gRPC context.
// If some of parameters has not been specified in query string of incoming
// HTTP testRequest corresponding fields in `query.PaginationRequest` structure will be set
// to nil.
// If provided pagination parameters are invalid function returns
// `status.Error(codes.InvalidArgument, parser_error)`
// See: `query.ParsePagination` for details.
func Pagination(ctx context.Context) (*query.Pagination, error) {
	l, lok := Header(ctx, limitMetaKey)
	o, ook := Header(ctx, offsetMetaKey)
	pt, ptok := Header(ctx, pageTokenMetaKey)

	if !lok && !ook && !ptok {
		return nil, nil
	}

	p, err := query.ParsePagination(l, o, pt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return p, nil
}

// SetPagination sets page info to outgoing gRPC context.
func SetPageInfo(ctx context.Context, p *query.PageInfo) error {
	m := make(map[string]string)

	if pt := p.GetPageToken(); pt != "" {
		m[pageInfoPageTokenMetaKey] = pt
	}

	if o := p.GetOffset(); o != 0 && p.NoMore() {
		m[pageInfoOffsetMetaKey] = "null"
	} else if o != 0 {
		m[pageInfoOffsetMetaKey] = strconv.FormatUint(uint64(o), 10)
	}

	if s := p.GetSize(); s != 0 {
		m[pageInfoSizeMetaKey] = strconv.FormatUint(uint64(s), 10)
	}

	return grpc.SetHeader(ctx, metadata.New(m))
}
