package service

import (
	my_errors "tender-service/internal/errors"
	"tender-service/internal/models"
	"tender-service/internal/repository"
)

type UserService interface {
	GetUserIDByUsername(username string) (string, error)
	GetUserByUsername(username string) (*models.User, error)
	CheckUserPermission(userID, organizationID string) (bool, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUserIDByUsername(username string) (string, error) {
	userID, err := s.repo.FindUserIDByUsername(username)
	if err != nil {
		if err == my_errors.ErrUserNotFound {
			return "", my_errors.ErrUserNotFound
		}
		return "", err
	}
	return userID, nil
}

func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		if err == my_errors.ErrUserNotFound {
			return nil, my_errors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) CheckUserPermission(userID, organizationID string) (bool, error) {
	hasPermission, err := s.repo.CheckUserPermission(userID, organizationID)
	if err != nil {
		return false, err
	}
	if !hasPermission {
		return false, my_errors.ErrForbidden
	}
	return true, nil
}
