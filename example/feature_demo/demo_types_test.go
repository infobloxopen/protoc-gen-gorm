package example

import (
	"context"
	"testing"

	"github.com/sbhagate-infoblox/protoc-gen-gorm/types"
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

func TestTimeOnlyToORM(t *testing.T) {
	t.Run("TimeOnly value to ORM", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got %v", r)
			}
		}()
		expectedTime := "01:59:18"
		pb := &TypeWithID{TimeOnly: &types.TimeOnly{Value: 7158}}
		orm, err := pb.ToORM(context.Background())
		if err != nil {
			t.Fatalf("failed to convert pb to orm: %v", err)
		}
		if orm.TimeOnly != expectedTime {
			t.Errorf("orm.TimeOnly= %v; want %s", orm.TimeOnly, expectedTime)
		}
	})
}

func TestTimeOnlyToPB(t *testing.T) {
	t.Run("TimeOnly value to PB", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got %v", r)
			}
		}()
		expectedTime := uint32(7158)
		orm := &TypeWithIDORM{TimeOnly: "01:59:18"}
		pb, err := orm.ToPB(context.Background())
		if err != nil {
			t.Fatalf("failed to convert pb to orm: %v", err)
		}
		if pb.TimeOnly == nil {
			t.Fatal("TimeOnly should not be nil")
		}
		if pb.TimeOnly.Value != expectedTime {
			t.Errorf("pb.TimeOnly.Value= %d; want %d", pb.TimeOnly.Value, expectedTime)
		}
	})
}

func TestTypeWithID_ToORM(t *testing.T) {
	t.Run("JoinTable", func(t *testing.T) {
		id := uint32(2468)
		maIDs := []uint32{1, 3, 5, 7, 9}
		pb := &TypeWithID{Id: id, MultiaccountTypeIds: maIDs}
		orm, err := pb.ToORM(context.Background())
		if err != nil {
			t.Fatalf("pb.ToORM=%v, want success", err)
		}
		if got, want := len(orm.MultiAccountTypes), len(maIDs); got != want {
			t.Errorf("len(orm.MultiAccountTypes) = %d; want %d", got, want)
		} else {
			for i, jt := range orm.MultiAccountTypes {
				if got, want := jt.TypeWithIDID, id; got != want {
					t.Errorf("jt.TypeWithIDID=%d; want %d", got, want)
				}
				if got, want := jt.MultiAccountTypeWithID, maIDs[i]; got != want {
					t.Errorf("jt.MultiAccountTypeWithID=%d; want %d", got, want)
				}
			}
		}
	})
}

func TestTypeWithIDORM_ToPB(t *testing.T) {
	t.Run("JoinTable", func(t *testing.T) {
		orm := &TypeWithIDORM{MultiAccountTypes: []*JoinTable{{MultiAccountTypeWithID: 1}, {MultiAccountTypeWithID: 3}}}
		pb, err := orm.ToPB(context.Background())
		if err != nil {
			t.Fatalf("orm.ToPB=%v; want success", err)
		}
		if got, want := len(pb.MultiaccountTypeIds), len(orm.MultiAccountTypes); got != want {
			t.Errorf("len(pb.MultiaccountTypeIds)=%d; want %d", got, want)
		} else {
			for i, maID := range pb.MultiaccountTypeIds {
				if got, want := maID, orm.MultiAccountTypes[i].MultiAccountTypeWithID; got != want {
					t.Errorf("pb.MultiaccountTypeIds[%d]=%d; want %d", i, got, want)
				}
			}
		}
	})
}