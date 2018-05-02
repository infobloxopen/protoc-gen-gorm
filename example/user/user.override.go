package user

import "time"

func (m *UserORM) AfterToPB(user *User) {
	user.Age = uint32(time.Now().Sub(m.Birthday).Hours() / 24 / 365)
}
