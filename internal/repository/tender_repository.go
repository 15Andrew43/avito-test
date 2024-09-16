package repository

import (
	"database/sql"
	my_errors "tender-service/internal/errors"
	"tender-service/internal/models"
)

type TenderRepository interface {
	GetTenders(serviceType string) ([]models.Tender, error)
	CreateTender(tender models.Tender) (models.Tender, error)
	GetTenderByID(tenderId string) (models.Tender, error)
	UpdateTenderStatus(tender models.Tender) error
	UpdateTender(tender models.Tender) error
	GetUserTenders(username string) ([]models.Tender, error)
	IsUserResponsibleForOrganization(userId, organizationId string) (bool, error)
	GetTenderHistoryVersion(tenderId string, version int) (models.TenderHistory, error)
	GetTenderHistoryByVersion(tenderId string, version int) (models.TenderHistory, error)
}

type tenderRepository struct {
	db *sql.DB
}

func NewTenderRepository(db *sql.DB) TenderRepository {
	return &tenderRepository{db: db}
}

func (r *tenderRepository) GetTenders(serviceType string) ([]models.Tender, error) {
	var tenders []models.Tender
	var rows *sql.Rows
	var err error

	if serviceType != "" {
		query := "SELECT id, name, description, service_type, status, organization_id, creator_id, version, created_at, updated_at FROM tender WHERE service_type = $1"
		rows, err = r.db.Query(query, serviceType)
	} else {
		query := "SELECT id, name, description, service_type, status, organization_id, creator_id, version, created_at, updated_at FROM tender"
		rows, err = r.db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tender models.Tender
		if err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.OrganizationID, &tender.CreatorID, &tender.Version, &tender.CreatedAt, &tender.UpdatedAt); err != nil {
			return nil, err
		}
		tenders = append(tenders, tender)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tenders, nil
}

func (r *tenderRepository) CreateTender(tender models.Tender) (models.Tender, error) {
	queryInsertTender := `
		INSERT INTO tender (name, description, service_type, status, organization_id, creator_id)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, name, description, service_type, status, organization_id, creator_id, version, created_at, updated_at
	`
	err := r.db.QueryRow(queryInsertTender, tender.Name, tender.Description, tender.ServiceType, tender.Status, tender.OrganizationID, tender.CreatorID).
		Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.OrganizationID, &tender.CreatorID, &tender.Version, &tender.CreatedAt, &tender.UpdatedAt)

	return tender, err
}

func (r *tenderRepository) GetUserTenders(username string) ([]models.Tender, error) {
	var tenders []models.Tender

	query := `
		SELECT t.id, t.name, t.description, t.service_type, t.status, t.organization_id, t.creator_id, t.version, t.created_at, t.updated_at
		FROM tender t
		JOIN employee e ON t.creator_id = e.id
		WHERE e.username = $1
	`
	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tender models.Tender
		if err := rows.Scan(
			&tender.ID,
			&tender.Name,
			&tender.Description,
			&tender.ServiceType,
			&tender.Status,
			&tender.OrganizationID,
			&tender.CreatorID,
			&tender.Version,
			&tender.CreatedAt,
			&tender.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tenders = append(tenders, tender)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tenders, nil
}

func (r *tenderRepository) GetTenderByID(tenderId string) (models.Tender, error) {
	var tender models.Tender
	query := "SELECT id, name, status, organization_id, creator_id FROM tender WHERE id = $1"
	err := r.db.QueryRow(query, tenderId).Scan(&tender.ID, &tender.Name, &tender.Status, &tender.OrganizationID, &tender.CreatorID)
	if err == sql.ErrNoRows {
		return tender, my_errors.ErrTenderNotFound
	} else if err != nil {
		return tender, err
	}
	return tender, nil
}
func (r *tenderRepository) UpdateTenderStatus(tender models.Tender) error {
	query := `UPDATE tender SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, tender.Status, tender.ID)
	return err
}

func (r *tenderRepository) UpdateTender(tender models.Tender) error {
	query := `
        UPDATE tender
        SET name = $1, description = $2, service_type = $3, updated_at = NOW()
        WHERE id = $4
    `
	_, err := r.db.Exec(query, tender.Name, tender.Description, tender.ServiceType, tender.ID)
	return err
}

func (r *tenderRepository) IsUserResponsibleForOrganization(userId, organizationId string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM organization_responsible WHERE user_id = $1 AND organization_id = $2)"
	err := r.db.QueryRow(query, userId, organizationId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *tenderRepository) GetTenderHistoryVersion(tenderId string, version int) (models.TenderHistory, error) {
	var history models.TenderHistory
	query := `SELECT tender_id, name, description, service_type, status, organization_id, creator_id, version, updated_at
              FROM tender_history
              WHERE tender_id = $1 AND version = $2`
	err := r.db.QueryRow(query, tenderId, version).Scan(
		&history.TenderID,
		&history.Name,
		&history.Description,
		&history.ServiceType,
		&history.Status,
		&history.OrganizationID,
		&history.CreatorID,
		&history.Version,
		&history.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return history, my_errors.ErrTenderHistoryNotFound
	} else if err != nil {
		return history, err
	}
	return history, nil
}

func (r *tenderRepository) GetTenderHistoryByVersion(tenderId string, version int) (models.TenderHistory, error) {
	query := `
        SELECT id, tender_id, name, description, service_type, status, organization_id, creator_id, version, updated_at
        FROM tender_history
        WHERE tender_id = $1 AND version = $2
    `
	var history models.TenderHistory
	err := r.db.QueryRow(query, tenderId, version).Scan(
		&history.ID,
		&history.TenderID,
		&history.Name,
		&history.Description,
		&history.ServiceType,
		&history.Status,
		&history.OrganizationID,
		&history.CreatorID,
		&history.Version,
		&history.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return history, my_errors.ErrTenderHistoryNotFound
		}
		return history, err
	}
	return history, nil
}
