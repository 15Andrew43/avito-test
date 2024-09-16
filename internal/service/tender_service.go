package service

import (
	"errors"
	"log"
	"tender-service/internal/models"
	"tender-service/internal/repository"

	my_errors "tender-service/internal/errors"

	"github.com/google/uuid"
)

type TenderService interface {
	GetTenders(serviceType string) ([]models.Tender, error)
	CreateTender(tender models.Tender, creatorUsername string) (models.Tender, error)
	GetUserTenders(username string) ([]models.Tender, error)
	GetTenderStatus(tenderId, username string) (models.TenderStatus, error)
	UpdateTenderStatus(tenderId string, status models.TenderStatus, username string) error
	EditTender(tenderId, username string, name, description, serviceType *string) error
	RollbackTenderVersion(tenderId string, version int, username string) error
}

type tenderService struct {
	repo        repository.TenderRepository
	userService UserService
}

func NewTenderService(repo repository.TenderRepository, userService UserService) TenderService {
	return &tenderService{repo: repo, userService: userService}
}

func (s *tenderService) GetTenders(serviceType string) ([]models.Tender, error) {
	return s.repo.GetTenders(serviceType)
}

func (s *tenderService) CreateTender(tender models.Tender, creatorUsername string) (models.Tender, error) {

	creatorID, err := s.userService.GetUserIDByUsername(creatorUsername)
	if err != nil {
		if errors.Is(err, my_errors.ErrUserNotFound) {
			return models.Tender{}, my_errors.ErrUnauthorized
		}
		return models.Tender{}, err
	}

	tender.CreatorID = creatorID

	createdTender, err := s.repo.CreateTender(tender)
	if err != nil {
		return models.Tender{}, err
	}

	return createdTender, nil
}

func (s *tenderService) GetUserTenders(username string) ([]models.Tender, error) {
	return s.repo.GetUserTenders(username)
}

func (s *tenderService) GetTenderStatus(tenderId, username string) (models.TenderStatus, error) {

	_, err := uuid.Parse(tenderId)
	if err != nil {
		return "", my_errors.ErrBadRequest
	}

	tender, err := s.repo.GetTenderByID(tenderId)
	if err != nil {
		if errors.Is(err, my_errors.ErrTenderNotFound) {
			return "", my_errors.ErrTenderNotFound
		}
		return "", err
	}

	userId, err := s.userService.GetUserIDByUsername(username)
	if err != nil {
		return "", my_errors.ErrUnauthorized
	}

	isResponsible, err := s.repo.IsUserResponsibleForOrganization(userId, tender.OrganizationID)
	if err != nil {
		return "", err
	}

	if tender.CreatorID != userId && !isResponsible {
		return "", my_errors.ErrForbidden
	}

	return tender.Status, nil
}

func (s *tenderService) UpdateTenderStatus(tenderId string, status models.TenderStatus, username string) error {

	_, err := uuid.Parse(tenderId)
	if err != nil {
		return my_errors.ErrBadRequest
	}

	tender, err := s.repo.GetTenderByID(tenderId)
	if err != nil {
		if errors.Is(err, my_errors.ErrTenderNotFound) {
			return my_errors.ErrTenderNotFound
		}
		return err
	}

	userId, err := s.userService.GetUserIDByUsername(username)
	if err != nil {
		return my_errors.ErrUnauthorized
	}

	isResponsible, err := s.repo.IsUserResponsibleForOrganization(userId, tender.OrganizationID)
	if err != nil {
		return err
	}

	if tender.CreatorID != userId && !isResponsible {
		return my_errors.ErrForbidden
	}

	tender.Status = status
	err = s.repo.UpdateTenderStatus(tender)
	if err != nil {
		return err
	}

	return nil
}

func (s *tenderService) EditTender(tenderId, username string, name, description, serviceType *string) error {
	_, err := uuid.Parse(tenderId)
	if err != nil {
		return my_errors.ErrBadRequest
	}

	tender, err := s.repo.GetTenderByID(tenderId)
	if err != nil {
		if errors.Is(err, my_errors.ErrTenderNotFound) {
			return my_errors.ErrTenderNotFound
		}
		return err
	}

	userId, err := s.userService.GetUserIDByUsername(username)
	if err != nil {
		return my_errors.ErrUnauthorized
	}

	isResponsible, err := s.repo.IsUserResponsibleForOrganization(userId, tender.OrganizationID)
	if err != nil {
		return err
	}

	if tender.CreatorID != userId && !isResponsible {
		return my_errors.ErrForbidden
	}

	if name != nil {
		tender.Name = *name
	}
	if description != nil {
		tender.Description = *description
	}
	if serviceType != nil {
		tender.ServiceType = *serviceType
	}

	err = s.repo.UpdateTender(tender)
	if err != nil {
		return err
	}

	return nil
}

func (s *tenderService) RollbackTenderVersion(tenderId string, version int, username string) error {
	log.Printf("RollbackTenderVersion: Parsing tender ID: %s", tenderId)
	_, err := uuid.Parse(tenderId)
	if err != nil {
		log.Printf("RollbackTenderVersion: Invalid tender ID format: %s", tenderId)
		return my_errors.ErrBadRequest
	}

	log.Printf("RollbackTenderVersion: Fetching tender by ID: %s", tenderId)
	tender, err := s.repo.GetTenderByID(tenderId)
	if err != nil {
		if errors.Is(err, my_errors.ErrTenderNotFound) {
			log.Printf("RollbackTenderVersion: Tender not found: %s", tenderId)
			return my_errors.ErrTenderNotFound
		}
		log.Printf("RollbackTenderVersion: Error fetching tender: %v", err)
		return err
	}

	log.Printf("RollbackTenderVersion: Fetching tender history for version: %d, tenderId: %s", version, tenderId)
	history, err := s.repo.GetTenderHistoryByVersion(tenderId, version)
	if err != nil {
		if errors.Is(err, my_errors.ErrTenderHistoryNotFound) {
			log.Printf("RollbackTenderVersion: Tender history not found for version: %d, tenderId: %s", version, tenderId)
			return my_errors.ErrTenderHistoryNotFound
		}
		log.Printf("RollbackTenderVersion: Error fetching tender history: %v", err)
		return err
	}

	log.Printf("RollbackTenderVersion: Fetching user ID by username: %s", username)
	userId, err := s.userService.GetUserIDByUsername(username)
	if err != nil {
		log.Printf("RollbackTenderVersion: Unauthorized user: %s", username)
		return my_errors.ErrUnauthorized
	}

	log.Printf("RollbackTenderVersion: Checking if user ID: %s is responsible for organization: %s", userId, tender.OrganizationID)
	isResponsible, err := s.repo.IsUserResponsibleForOrganization(userId, tender.OrganizationID)
	if err != nil {
		log.Printf("RollbackTenderVersion: Error checking user responsibility: %v", err)
		return err
	}

	if tender.CreatorID != userId && !isResponsible {
		log.Printf("RollbackTenderVersion: Forbidden access for user: %s on tender ID: %s", username, tenderId)
		return my_errors.ErrForbidden
	}

	log.Printf("RollbackTenderVersion: Rolling back tender to version: %d", version)
	tender.Name = history.Name
	tender.Description = history.Description
	tender.ServiceType = history.ServiceType
	tender.Status = history.Status

	log.Printf("RollbackTenderVersion: Updating tender with new values")
	err = s.repo.UpdateTender(tender)
	if err != nil {
		log.Printf("RollbackTenderVersion: Error updating tender: %v", err)
		return err
	}

	log.Printf("RollbackTenderVersion: Successfully rolled back tender ID: %s to version: %d", tenderId, version)
	return nil
}
