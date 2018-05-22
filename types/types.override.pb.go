package types

// XXX_WellKnownType is a hack -- please don't make a pattern out of this!
// This is used to trick certain toolsets to treat this field like a WKT StringValue.
// See https://github.com/golang/protobuf/blob/70c277a8a150a8e069492e6600926300405c2884/jsonpb/jsonpb.go#L157
// and https://github.com/golang/protobuf/blob/70c277a8a150a8e069492e6600926300405c2884/jsonpb/jsonpb.go#L191
func (*UUIDValue) XXX_WellKnownType() string { return "StringValue" }
