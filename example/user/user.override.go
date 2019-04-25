package user

import (
	"context"
	"time"
)

// AfterToPB implements the posthook interface for the User type. This allows
// us to customize conversion behavior. In this example, we set the User's Age
// based on the Birthday, instead of storing it separately in the DB
func (m *UserORM) AfterToPB(ctx context.Context, user *User) error {
	user.Age = uint32(time.Now().Sub(*m.Birthday).Hours() / 24 / 365)
	return nil
}
