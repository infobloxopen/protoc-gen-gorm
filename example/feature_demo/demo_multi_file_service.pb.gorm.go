// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/demo_multi_file_service.proto

package example

import context "context"

import gorm1 "github.com/jinzhu/gorm"
import json1 "encoding/json"
import trace1 "go.opencensus.io/trace"

import fmt "fmt"
import math "math"
import _ "github.com/infobloxopen/atlas-app-toolkit/query"

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math.Inf

type BlogPostServiceDefaultServer struct {
	DB *gorm1.DB
}

// spanCreate ...
func (m *BlogPostServiceDefaultServer) spanCreate(ctx context.Context, in interface{}, methodName string) (*trace1.Span, error) {
	_, span := trace1.StartSpan(ctx, field_mask1Sprint("BlogPostServiceDefaultServer.", methodName))
	rawParameter, errMarshaling := json1.Marshal(in)
	if errMarshaling != nil {
		return nil, errMarshaling
	}
	span.Annotate([]trace1.Attribute{trace1.StringAttribute("in", string(rawParameter))}, "in parameter")
	return span, nil
}

// spanError ...
func (m *BlogPostServiceDefaultServer) spanError(span *trace1Span, err error) error {
	span.SetStatus(trace.Status{
		Code:    trace1.StatusCodeUnknown,
		Message: err.Error(),
	})
	return err
}

// spanResult ...
func (m *BlogPostServiceDefaultServer) spanResult(span *trace1Span, out interface{}) error {
	rawParameter, errMarshaling := json1.Marshal(out)
	if errMarshaling != nil {
		return errMarshaling
	}
	span.Annotate([]trace1.Attribute{trace1.StringAttribute("out", string(rawParameter))}, "out parameter")
}

// Read ...
func (m *BlogPostServiceDefaultServer) Read(ctx context.Context, in *ReadAccountRequest) (*ReadBlogPostsResponse, error) {
	span, errSpanCreate := m.spanCreate(ctx, in, "Read")
	if errSpanCreate != nil {
		return nil, errSpanCreate
	}
	defer span.End()
	out := &ReadBlogPostsResponse{}
	err = m.spanResult(span, out)
	if err != nil {
		return nil, m.spanError(span, err)
	}
	return out, nil
}
