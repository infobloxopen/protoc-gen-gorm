package gormable_types

// XXX_WellKnownType is a hack -- please don't make a pattern out of this! This is used to trick certain toolsets to treat this field like a WKT StringValue.
func (*UUIDValue) XXX_WellKnownType() string { return "StringValue" }
