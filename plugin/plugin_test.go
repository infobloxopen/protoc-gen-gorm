package plugin

import (
	"reflect"
	"testing"
)

func TestGetOrmable(t *testing.T) {
	tests := []struct {
		in  string
		m   map[string]*OrmableType
		e   *OrmableType
		err error
	}{
		{
			in: "IntPoint",
			m: map[string]*OrmableType{
				"IntPoint": NewOrmableType("", "google.protobuf", nil),
			},
			e: NewOrmableType("", "google.protobuf", nil),
		},
		{
			in: "Task",
			m: map[string]*OrmableType{
				"Task": NewOrmableType("TaskORM", "google.protobuf", nil),
			},
			e: NewOrmableType("TaskORM", "google.protobuf", nil),
		},
	}

	for _, tt := range tests {
		ot, err := GetOrmable(tt.m, tt.in)
		if err != tt.err {
			t.Errorf("got: %s wanted: %s", err, tt.err)
		}

		if !reflect.DeepEqual(ot, tt.e) {
			t.Errorf("got: %+v wanted: %+v", *ot, tt.e)
		}
	}
}
