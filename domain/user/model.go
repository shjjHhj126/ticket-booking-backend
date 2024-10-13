package user

// type User struct {
// 	ID             int    `db:"id"`
// 	Username       string `db:"username"`
// 	Email          string `db:"email"`
// 	HashedPassword []byte `db:"password_hash"`
// }

// func (postUser userApi.PostUser) reqToModel() User, error {
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return User{
// 		Username:       postUser.Username,
// 		Email:          postUser.Email,
// 		HashedPassword: hashedPassword,
// 	}, nil
// }
