package user

import "github.com/jmoiron/sqlx"

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Create(user *User) error {

	query := "INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)"

	repo.db.Exec(query, user.Username, user.Email)

	return nil
}
