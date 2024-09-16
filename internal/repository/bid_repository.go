package repository

import (
	"database/sql"
	"fmt"
	"tender-service/internal/models"

	my_errors "tender-service/internal/errors"
)

type BidRepository interface {
	CreateBid(bid *models.Bid) (*models.Bid, error)
	GetBidsByTenderID(tenderID string, limit, offset int) ([]models.Bid, error)
	GetBidsByUserID(userID string, limit, offset int) ([]models.Bid, error)
	GetBidByID(bidID string) (*models.Bid, error)
	UpdateBidStatus(bidID string, status models.BidStatus) error
	EditBid(bidID string, updates map[string]interface{}) error
	AddBidFeedback(bidID, feedback string) error
}

type bidRepository struct {
	db *sql.DB
}

func NewBidRepository(db *sql.DB) BidRepository {
	return &bidRepository{db: db}
}

func (r *bidRepository) CreateBid(bid *models.Bid) (*models.Bid, error) {
	query := `
        INSERT INTO bid (description, tender_id, organization_id, user_id, author_type, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        RETURNING id, description, tender_id, organization_id, user_id, author_type, status, created_at, updated_at
    `
	err := r.db.QueryRow(query, bid.Description, bid.TenderID, bid.OrganizationID, bid.UserID, bid.AuthorType, bid.Status).
		Scan(&bid.ID, &bid.Description, &bid.TenderID, &bid.OrganizationID, &bid.UserID, &bid.AuthorType, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return bid, nil
}

func (r *bidRepository) GetBidsByTenderID(tenderID string, limit, offset int) ([]models.Bid, error) {
	var bids []models.Bid

	query := `
        SELECT id, description, tender_id, organization_id, user_id, status, created_at, updated_at 
        FROM bid 
        WHERE tender_id = $1 
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(query, tenderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid models.Bid
		if err := rows.Scan(&bid.ID, &bid.Description, &bid.TenderID, &bid.OrganizationID, &bid.UserID, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, nil
}

func (r *bidRepository) GetBidsByUserID(userID string, limit, offset int) ([]models.Bid, error) {
	query := `
        SELECT id, description, tender_id, organization_id, user_id, status, created_at, updated_at
        FROM bid
        WHERE user_id = $1
        ORDER BY description
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var bid models.Bid
		if err := rows.Scan(&bid.ID, &bid.Description, &bid.TenderID, &bid.OrganizationID, &bid.UserID, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, nil
}

func (r *bidRepository) GetUserBids(userID string, limit, offset int) ([]models.Bid, error) {
	query := `
		SELECT id, tender_id, organization_id, user_id, author_type, description, status, created_at, updated_at
		FROM bid
		WHERE user_id = $1
		ORDER BY description ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var bid models.Bid
		if err := rows.Scan(&bid.ID, &bid.TenderID, &bid.OrganizationID, &bid.UserID, &bid.AuthorType, &bid.Description, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}
	return bids, nil
}

func (r *bidRepository) GetBidByID(bidID string) (*models.Bid, error) {
	query := `SELECT id, tender_id, organization_id, user_id, description, status, author_type FROM bid WHERE id = $1`
	row := r.db.QueryRow(query, bidID)

	var bid models.Bid
	err := row.Scan(&bid.ID, &bid.TenderID, &bid.OrganizationID, &bid.UserID, &bid.Description, &bid.Status, &bid.AuthorType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, my_errors.ErrBidNotFound
		}
		return nil, err
	}
	return &bid, nil
}

func (r *bidRepository) UpdateBidStatus(bidID string, status models.BidStatus) error {
	query := `
        UPDATE bid 
        SET status = $1, updated_at = NOW() 
        WHERE id = $2
        RETURNING id
    `
	var id string
	err := r.db.QueryRow(query, status, bidID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return my_errors.ErrBidNotFound
		}
		return err
	}
	return nil
}

func (r *bidRepository) EditBid(bidID string, updates map[string]interface{}) error {
	query := `UPDATE bid SET `
	var params []interface{}
	paramIndex := 1

	for field, value := range updates {
		if paramIndex > 1 {
			query += ", "
		}
		query += field + " = $" + fmt.Sprintf("%d", paramIndex)
		params = append(params, value)
		paramIndex++
	}

	query += ", updated_at = NOW() WHERE id = $" + fmt.Sprintf("%d", paramIndex)
	params = append(params, bidID)

	result, err := r.db.Exec(query, params...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return my_errors.ErrBidNotFound
	}

	return nil
}

func (r *bidRepository) AddBidFeedback(bidID, feedback string) error {
	query := `
        INSERT INTO bid_review (bid_id, description, created_at)
        VALUES ($1, $2, NOW())
    `
	_, err := r.db.Exec(query, bidID, feedback)
	if err != nil {
		return err
	}
	return nil
}
