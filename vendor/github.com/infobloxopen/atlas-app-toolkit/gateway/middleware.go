package gateway

import (
	"context"
	"fmt"
	"reflect"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware if an user's testRequest message
// defines any of collection operators.
//
// Returned middleware populates collection operators from gRPC metadata if
// they defined in a testRequest message.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		// handle panic
		defer func() {
			if perr := recover(); perr != nil {
				err = status.Errorf(codes.Internal, "collection operators interceptor: %s", perr)
				grpclog.Errorln(err)
				res, err = nil, err
			}
		}()

		if req == nil {
			grpclog.Warningf("collection operator interceptor: empty testRequest %+v", req)
			return handler(ctx, req)
		}

		// looking for op.Sorting
		sorting, err := Sorting(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid sorting operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if sorting != nil {
			if err := setOp(req, sorting); err != nil {
				grpclog.Errorf("collection operator interceptor: failed to set sorting operator - %s", err)
			}
		}

		// looking for op.FieldSelection
		fieldSelection := FieldSelection(ctx)
		if fieldSelection != nil {
			if err := setOp(req, fieldSelection); err != nil {
				grpclog.Errorf("collection operator interceptor: failed to set field selection operator - %s", err)
			}
		}

		// looking for op.Filtering
		filtering, err := Filtering(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid filtering operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if filtering != nil {
			if err := setOp(req, filtering); err != nil {
				grpclog.Errorf("collection operator interceptor: failed to set filtering operator - %s", err)
			}
		}

		// looking for op.ClientDrivenPagination
		pagination, err := Pagination(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid pagination operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if pagination != nil {
			if err := setOp(req, pagination); err != nil {
				grpclog.Errorf("collection operator interceptor: failed to set pagination operator - %s", err)
			}
		}

		res, err = handler(ctx, req)
		if err != nil {
			return res, err
		}

		// looking for op.PageInfo
		page := new(query.PageInfo)
		if err := unsetOp(res, page); err != nil {
			grpclog.Errorf("collection operator interceptor: failed to set page info - %s", err)
		}

		if err := SetPageInfo(ctx, page); err != nil {
			grpclog.Errorf("collection operator interceptor: failed to set page info - %s", err)
			return nil, err
		}

		return
	}
}

func setOp(req, op interface{}) error {
	reqval := reflect.ValueOf(req)

	if reqval.Kind() != reflect.Ptr {
		return fmt.Errorf("testRequest is not a pointer - %s", reqval.Kind())
	}

	reqval = reqval.Elem()

	if reqval.Kind() != reflect.Struct {
		return fmt.Errorf("testRequest value is not a struct - %s", reqval.Kind())
	}

	for i := 0; i < reqval.NumField(); i++ {
		f := reqval.FieldByIndex([]int{i})

		if f.Type() != reflect.TypeOf(op) {
			continue
		}

		if !f.IsValid() || !f.CanSet() {
			return fmt.Errorf("operation field %+v in testRequest %+v is invalid or cannot be set", op, req)
		}

		if vop := reflect.ValueOf(op); vop.IsValid() {
			f.Set(vop)
		}
	}

	return nil
}

func unsetOp(res, op interface{}) error {
	resval := reflect.ValueOf(res)
	if resval.Kind() != reflect.Ptr {
		return fmt.Errorf("response is not a pointer - %s", resval.Kind())
	}

	resval = resval.Elem()
	if resval.Kind() != reflect.Struct {
		return fmt.Errorf("response value is not a struct - %s", resval.Kind())
	}

	opval := reflect.ValueOf(op)
	if opval.Kind() != reflect.Ptr {
		return fmt.Errorf("operator is not a pointer - %s", opval.Kind())
	}

	for i := 0; i < resval.NumField(); i++ {
		f := resval.FieldByIndex([]int{i})

		if f.Type() != opval.Type() {
			continue
		}

		if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.Ptr {
			return fmt.Errorf("operation field %T in response %+v is invalid or cannot be set", op, res)
		}

		if o := opval.Elem(); o.IsValid() && o.CanSet() && f.Elem().IsValid() {
			o.Set(f.Elem())
		}

		f.Set(reflect.Zero(f.Type()))
	}

	return nil
}
