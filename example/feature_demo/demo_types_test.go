package example

import (
	"context"
	"testing"
)

func TestInet(t *testing.T) {
	// a nil Inet originally panicked -- this test ensures that the panic is fixed
	t.Run("nil Inet value doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got %v", r)
			}
		}()
		pb := &TypeWithID{}
		orm, err := pb.ToORM(context.Background())
		if err != nil {
			t.Fatalf("failed to convert pb to orm: %v", err)
		}
		if orm.Address != nil {
			t.Errorf("orm.Address= %v; want nil", orm.Address)
		}
	})
}
