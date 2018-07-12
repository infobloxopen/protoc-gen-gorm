package example

import "context"

type JoinTable struct {
	TypeWithIDID           uint32
	MultiAccountTypeWithID uint32
}

func (pb *TypeWithID) AfterToORM(ctx context.Context, orm *TypeWithIDORM) error {
	var maIDs []*JoinTable
	for _, id := range pb.MultiaccountTypeIds {
		maIDs = append(maIDs, &JoinTable{TypeWithIDID: orm.Id, MultiAccountTypeWithID: id})
	}
	orm.MultiAccountTypes = maIDs
	return nil
}

func (orm *TypeWithIDORM) AfterToPB(ctx context.Context, pb *TypeWithID) error {
	var ids []uint32
	for _, jt := range orm.MultiAccountTypes {
		ids = append(ids, jt.MultiAccountTypeWithID)
	}
	pb.MultiaccountTypeIds = ids

	return nil
}
