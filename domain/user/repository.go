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
