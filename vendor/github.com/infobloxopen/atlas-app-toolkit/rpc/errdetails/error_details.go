package errdetails

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
)

// New returns a TargetInfo representing c, target and msg.
// Converts provided Code to int32
func New(c codes.Code, target string, msg string) *TargetInfo {
	return &TargetInfo{Code: int32(c), Message: msg, Target: target}
}

// NewfTargetInfo returns NewTargetInfo(c, fmt.Sprintf(format, a...)).
func Newf(c codes.Code, target string, format string, a ...interface{}) *TargetInfo {
	return New(c, target, fmt.Sprintf(format, a...))
}

// MarshalJSON implements json.Marshaler.
// TargetInfo.Code field is marshaled into string with corresponding value,
// see [google.rpc.Code][google.rpc.Code], if code is the codes.Unimplemented
// it is marshaled as "NOT_IMPLEMENTED" string.
func (ti *TargetInfo) MarshalJSON() ([]byte, error) {
	v := make(map[string]string, 3)
	if m := ti.GetMessage(); m != "" {
		v["message"] = ti.Message
	}
	if t := ti.GetTarget(); t != "" {
		v["target"] = ti.Target
	}
	if ti.GetCode() == int32(codes.Unimplemented) {
		v["code"] = "NOT_IMPLEMENTED"
	} else {
		v["code"] = code.Code(ti.GetCode()).String()
	}

	return json.Marshal(&v)
}

// UnmarshalJSON implements json.Unmarshaler.
// If "code" is not provided in JSON data or is null,
// the TargetInfo.Code will be set to 0 (OK)
func (ti *TargetInfo) UnmarshalJSON(data []byte) error {
	v := make(map[string]string, 3)
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	ti.Message = v["message"]
	ti.Target = v["target"]

	// if not provided or null
	if s, ok := v["code"]; !ok || s == "" {
		ti.Code = int32(code.Code_OK)
	} else if c, ok := code.Code_value[strings.ToUpper(s)]; ok {
		ti.Code = c
	} else if strings.ToUpper(s) == "NOT_IMPLEMENTED" {
		ti.Code = int32(code.Code_UNIMPLEMENTED)
	} else {
		ti.Code = int32(code.Code_UNKNOWN)
	}
	return nil
}
