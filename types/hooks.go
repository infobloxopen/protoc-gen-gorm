package gormable_types

// WithBeforeToOrm has a hook function to be called before an Ormable type
// is converted ToOrm
type WithBeforeToOrm interface {
	BeforeToOrmHook(interface{})
}

// WithAfterToOrm has a hook function to be called after an Ormable type
// is converted ToOrm
type WithAfterToOrm interface {
	AfterToOrmHook(interface{})
}

// WithBeforeToPB has a hook function to be called before a Protobufable
// type is converted from orm ToPB
type WithBeforeToPB interface {
	BeforeToPBHook(interface{})
}

// WithAfterToPB has a hook function to be called right after a Protobufable
// type is converted from orm ToPB
type WithAfterToPB interface {
	AfterToPBHook(interface{})
}
