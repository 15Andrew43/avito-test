package repository

import (
	"database/sql"
	"errors"
	my_errors "tender-service/internal/errors"
	"tender-service/internal/models"
)

type UserRepository interface {
	FindUserIDByUsername(username string) (string, error)
	GetUserByID(userID string) (*models.User, error)
	CheckUserPermission(userID, organizationID string) (bool, error)
	GetUserByUsername(username string) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindUserIDByUsername(username string) (string, error) {
	var userID string
	query := "SELECT id FROM employee WHERE username = $1"
	err := r.db.QueryRow(query, username).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", my_errors.ErrUserNotFound
	} else if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *userRepository) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	query := "SELECT id, username, first_name, last_name, created_at, updated_at FROM employee WHERE id = $1"
	err := r.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, my_errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CheckUserPermission(userID, organizationID string) (bool, error) {
	var count int
	query := "SELECT COUNT(1) FROM organization_responsible WHERE user_id = $1 AND organization_id = $2"
	err := r.db.QueryRow(query, userID, organizationID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, first_name, last_name FROM employee WHERE username = $1`
	row := r.db.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, my_errors.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
