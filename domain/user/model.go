package user

import (
	"ticket-booking-backend/dto"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int    `db:"id"`
	Username       string `db:"username"`
	Email          string `db:"email"`
	HashedPassword []byte `db:"password_hash"`
}

func DtoToModel(postUser dto.PostUser) (User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(postUser.Password), 12)
	if err != nil {
		return User{}, err
	}
	return User{
		Username:       postUser.Username,
		Email:          postUser.Email,
		HashedPassword: hashedPassword,
	}, nil
}
