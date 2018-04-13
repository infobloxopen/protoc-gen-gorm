package gw

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/op"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//SetFieldSelection sets op.FieldSelection to gRPC metadata
func SetFieldSelection(ctx context.Context, fields *op.FieldSelection) error {
	fieldsStr := fields.GoString()
	md := metadata.Pairs(
		runtime.MetadataPrefix+fieldsMetaKey, fieldsStr,
	)
	return grpc.SetHeader(ctx, md)
}

//FieldSelection extracts op.FieldSelection from gRPC metadata
func FieldSelection(ctx context.Context) *op.FieldSelection {
	fields, ok := Header(ctx, fieldsMetaKey)
	if !ok {
		return nil
	}
	return op.ParseFieldSelection(fields)
}

//retainFields function extracts the configuration for fields that
//need to be ratained either from gRPC response or from original request
//(in case when gRPC side didn't set any preferences) and retains only
//this fields on outgoing response (dynmap).
func retainFields(ctx context.Context, req *http.Request, dynmap map[string]interface{}) {
	fieldsStr, ok := Header(ctx, fieldsMetaKey)
	if !ok && req != nil {
		//no fields in gprc response -> try to get from original request
		vals := req.URL.Query()
		fieldsStr = vals.Get(fieldsQueryKey)
	}

	if fieldsStr == "" {
		return
	}

	fields := op.ParseFieldSelection(fieldsStr)
	if fields != nil {
		for _, result := range dynmap {
			if results, ok := result.([]interface{}); ok {
				for _, r := range results {
					if m, ok := r.(map[string]interface{}); ok {
						doRetainFields(m, fields.Fields)
					}
				}
			}
		}
	}
}

func doRetainFields(obj map[string]interface{}, fields op.FieldSelectionMap) {
	if fields == nil || len(fields) == 0 {
		return
	}

	for key := range obj {
		if _, ok := fields[key]; !ok {
			delete(obj, key)
		} else {
			switch x := obj[key].(type) {
			case map[string]interface{}:
				fds := fields[key].Subs
				if fds != nil && len(fds) > 0 {
					doRetainFields(x, fds)
				}
			case []interface{}:
				//ingnoring arrays for now
			}
		}
	}
}
