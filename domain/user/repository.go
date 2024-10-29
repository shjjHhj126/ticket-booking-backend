package user

import (
	"fmt"

	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Create(user *User) error {

	query := "INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)"

	_, err := repo.db.Exec(query, user.Username, user.Email, user.HashedPassword)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (repo *UserRepository) GetUserByEmail(email string) (User, error) {
	query := `
		SELECT id, username, password_hash, email
		FROM users
		WHERE email = $1
	`
	user := User{}
	err := repo.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.HashedPassword, &user.Email)
	if err != nil {
		return user, err
	}
	return user, nil
}
