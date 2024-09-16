package service

import (
	"database/sql"
	"errors"
	"log"
	my_errors "tender-service/internal/errors"
	"tender-service/internal/models"
	"tender-service/internal/repository"

	"github.com/google/uuid"
)

type BidService interface {
	CreateBid(description, tenderID, organizationID, userID string, authorType models.BidAuthorType) (*models.Bid, error)
	GetBidsByTenderID(tenderID, username string, limit, offset int) ([]models.Bid, error)
	GetUserBids(userID string, limit, offset int) ([]models.Bid, error)
	GetBidStatus(bidID string, username string) (models.BidStatus, error)
	UpdateBidStatus(bidID, status, username string) error
	EditBid(bidID, username string, updates map[string]interface{}) error
	SubmitBidFeedback(bidID, username, feedback string) (*models.Bid, error)
}

type bidService struct {
	repo       repository.BidRepository
	tenderRepo repository.TenderRepository
	userRepo   repository.UserRepository
}

func NewBidService(repo repository.BidRepository, tenderRepo repository.TenderRepository, userRepo repository.UserRepository) BidService {
	return &bidService{repo: repo, tenderRepo: tenderRepo, userRepo: userRepo}
}

func (s *bidService) CreateBid(description, tenderID, organizationID, userID string, authorType models.BidAuthorType) (*models.Bid, error) {
	if _, err := uuid.Parse(tenderID); err != nil {
		log.Printf("Invalid tenderID format: %s", tenderID)
		return nil, my_errors.ErrBadRequest
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		log.Printf("Invalid organizationID format: %s", organizationID)
		return nil, my_errors.ErrBadRequest
	}

	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("Invalid userID format: %s", userID)
		return nil, my_errors.ErrBadRequest
	}

	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, my_errors.ErrUserNotFound) {
			log.Printf("User with ID %s not found", userID)
			return nil, my_errors.ErrUnauthorized
		}
		return nil, err
	}

	_, err = s.tenderRepo.GetTenderByID(tenderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error finding tender with ID: %s, error: %v", tenderID, "tender not found")
			return nil, my_errors.ErrTenderNotFound
		}
		return nil, err
	}

	hasPermission, err := s.userRepo.CheckUserPermission(userID, organizationID)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		log.Printf("User %s does not have permission to bid for organization %s", userID, organizationID)
		return nil, my_errors.ErrForbidden
	}

	bid := &models.Bid{
		Description:    description,
		TenderID:       tenderID,
		OrganizationID: organizationID,
		UserID:         userID,
		AuthorType:     authorType,
		Status:         models.BidStatusCreated,
	}

	createdBid, err := s.repo.CreateBid(bid)
	if err != nil {
		return nil, err
	}

	return createdBid, nil
}

func (s *bidService) GetUserBids(username string, limit, offset int) ([]models.Bid, error) {
	log.Printf("GetUserBids called for username: %s, limit: %d, offset: %d", username, limit, offset)

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("Error retrieving user by username: %s, error: %v", username, err)
		if errors.Is(err, my_errors.ErrUserNotFound) {
			return nil, my_errors.ErrUserNotFound
		}
		return nil, err
	}

	bids, err := s.repo.GetBidsByUserID(user.ID, limit, offset)
	if err != nil {
		log.Printf("Error retrieving bids for user: %s, error: %v", username, err)
		return nil, err
	}

	log.Printf("Successfully retrieved bids for user: %s", username)
	return bids, nil
}

func (s *bidService) GetBidsByTenderID(tenderID, username string, limit, offset int) ([]models.Bid, error) {

	log.Printf("GetBidsByTenderID called with tenderID: %s, username: %s, limit: %d, offset: %d", tenderID, username, limit, offset)

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, my_errors.ErrUserNotFound) {
			log.Printf("User not found: %s", username)
		} else {
			log.Printf("Error retrieving user by username: %s, error: %v", username, err)
		}
		return nil, my_errors.ErrUserNotFound
	}
	log.Printf("User %s found with ID: %s", username, user.ID)

	_, err = uuid.Parse(tenderID)
	if err != nil {
		log.Printf("Invalid tender ID format: %s", tenderID)
		return nil, my_errors.ErrBadRequest
	}

	tender, err := s.tenderRepo.GetTenderByID(tenderID)
	if err != nil {
		if errors.Is(err, my_errors.ErrTenderNotFound) {
			log.Printf("Tender not found: %s", tenderID)
			return nil, my_errors.ErrTenderNotFound
		}
		log.Printf("Error retrieving tender by ID: %s, error: %v", tenderID, err)
		return nil, err
	}
	log.Printf("Tender %s found, associated with organization %s", tenderID, tender.OrganizationID)

	hasPermission, err := s.userRepo.CheckUserPermission(user.ID, tender.OrganizationID)
	if err != nil {
		log.Printf("Error checking user permission for user %s and organization %s: %v", user.ID, tender.OrganizationID, err)
		return nil, err
	}
	if !hasPermission {
		log.Printf("User %s does not have permission to access organization %s", user.ID, tender.OrganizationID)
		return nil, my_errors.ErrForbidden
	}
	log.Printf("User %s has permission to access organization %s", user.ID, tender.OrganizationID)

	bids, err := s.repo.GetBidsByTenderID(tenderID, limit, offset)
	if err != nil {
		if errors.Is(err, my_errors.ErrTenderNotFound) {
			log.Printf("Tender not found: %s", tenderID)
			return nil, my_errors.ErrTenderNotFound
		}
		log.Printf("Error retrieving bids for tender %s: %v", tenderID, err)
		return nil, err
	}

	log.Printf("Retrieved %d bids for tender %s", len(bids), tenderID)
	return bids, nil
}

func (s *bidService) GetBidStatus(bidID string, username string) (models.BidStatus, error) {
	log.Printf("GetBidStatus: Parsing bidID=%s", bidID)
	_, err := uuid.Parse(bidID)
	if err != nil {
		log.Printf("GetBidStatus: Invalid bidID format: %s", bidID)
		return "", my_errors.ErrBadRequest
	}

	log.Printf("GetBidStatus: Fetching bid by ID=%s", bidID)
	bid, err := s.repo.GetBidByID(bidID)
	if err != nil {
		log.Printf("GetBidStatus: Error fetching bid: %v", err)
		return "", err
	}

	log.Printf("GetBidStatus: Fetching user by username=%s", username)
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("GetBidStatus: User not found for username=%s", username)
		return "", my_errors.ErrUserNotFound
	}

	log.Printf("GetBidStatus: Checking access rights for username=%s on bidID=%s", username, bidID)
	if bid.UserID != user.ID {
		log.Printf("GetBidStatus: Access denied for username=%s on bidID=%s", username, bidID)
		return "", my_errors.ErrForbidden
	}

	log.Printf("GetBidStatus: Returning status=%s for bidID=%s", bid.Status, bidID)
	return bid.Status, nil
}

func (s *bidService) UpdateBidStatus(bidID, status, username string) error {
	log.Printf("UpdateBidStatus: Parsing bidID=%s", bidID)
	_, err := uuid.Parse(bidID)
	if err != nil {
		log.Printf("UpdateBidStatus: Invalid bidID format: %s", bidID)
		return my_errors.ErrBadRequest
	}

	log.Printf("UpdateBidStatus: Fetching bid by ID=%s", bidID)
	bid, err := s.repo.GetBidByID(bidID)
	if err != nil {
		log.Printf("UpdateBidStatus: Error fetching bid: %v", err)
		return err
	}

	log.Printf("UpdateBidStatus: Fetching user by username=%s", username)
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("UpdateBidStatus: User not found for username=%s", username)
		return my_errors.ErrUserNotFound
	}

	log.Printf("UpdateBidStatus: Checking access rights for username=%s on bidID=%s", username, bidID)
	if bid.UserID != user.ID {
		log.Printf("UpdateBidStatus: Access denied for username=%s on bidID=%s", username, bidID)
		return my_errors.ErrForbidden
	}

	bidStatus := models.BidStatus(status)

	log.Printf("UpdateBidStatus: Updating bid status to %s", bidStatus)
	err = s.repo.UpdateBidStatus(bidID, bidStatus)
	if err != nil {
		log.Printf("UpdateBidStatus: Error updating bid status: %v", err)
		return err
	}

	return nil
}

func (s *bidService) EditBid(bidID, username string, updates map[string]interface{}) error {
	log.Printf("EditBid: Parsing bidID=%s", bidID)
	_, err := uuid.Parse(bidID)
	if err != nil {
		log.Printf("EditBid: Invalid bidID format: %s", bidID)
		return my_errors.ErrBadRequest
	}

	log.Printf("EditBid: Fetching bid by ID=%s", bidID)
	bid, err := s.repo.GetBidByID(bidID)
	if err != nil {
		log.Printf("EditBid: Error fetching bid: %v", err)
		return err
	}

	log.Printf("EditBid: Fetching user by username=%s", username)
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("EditBid: User not found for username=%s", username)
		return my_errors.ErrUserNotFound
	}

	log.Printf("EditBid: Checking access rights for username=%s on bidID=%s", username, bidID)
	if bid.UserID != user.ID {
		log.Printf("EditBid: Access denied for username=%s on bidID=%s", username, bidID)
		return my_errors.ErrForbidden
	}

	log.Printf("EditBid: Applying updates to bid")
	err = s.repo.EditBid(bidID, updates)
	if err != nil {
		log.Printf("EditBid: Error updating bid: %v", err)
		return err
	}

	return nil
}

func (s *bidService) SubmitBidFeedback(bidID, username, feedback string) (*models.Bid, error) {
	log.Printf("SubmitBidFeedback: Fetching bid by ID=%s", bidID)

	bid, err := s.repo.GetBidByID(bidID)
	if err != nil {
		log.Printf("SubmitBidFeedback: Error fetching bid: %v", err)
		return nil, err
	}

	log.Printf("SubmitBidFeedback: Fetching user by username=%s", username)
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("SubmitBidFeedback: User not found for username=%s", username)
		return nil, my_errors.ErrUserNotFound
	}

	if bid.UserID != user.ID {
		log.Printf("SubmitBidFeedback: Access denied for username=%s on bidID=%s", username, bidID)
		return nil, my_errors.ErrForbidden
	}

	log.Printf("SubmitBidFeedback: Adding feedback for bidID=%s", bidID)
	err = s.repo.AddBidFeedback(bidID, feedback)
	if err != nil {
		log.Printf("SubmitBidFeedback: Error adding feedback: %v", err)
		return nil, err
	}

	log.Printf("SubmitBidFeedback: Feedback successfully added for bidID=%s", bidID)
	return bid, nil
}
