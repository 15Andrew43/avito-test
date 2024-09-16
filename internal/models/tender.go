package models

import (
	"errors"
	"time"
)

type TenderStatus string

const (
	Created   TenderStatus = "CREATED"
	Published TenderStatus = "PUBLISHED"
	Closed    TenderStatus = "CLOSED"
)

type Tender struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	ServiceType    string       `json:"serviceType"`
	Status         TenderStatus `json:"status"`
	OrganizationID string       `json:"organizationId"`
	CreatorID      string       `json:"creatorId"`
	Version        int          `json:"version"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type ErrorResponse struct {
	Reason string `json:"reason"`
}

type TenderHistory struct {
	ID             string       `json:"id"`
	TenderID       string       `json:"tender_id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	ServiceType    string       `json:"service_type"`
	Status         TenderStatus `json:"status"`
	OrganizationID string       `json:"organization_id"`
	CreatorID      string       `json:"creator_id"`
	Version        int          `json:"version"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

func ParseTenderStatus(status string) (TenderStatus, error) {
	switch status {
	case string(Created):
		return Created, nil
	case string(Published):
		return Published, nil
	case string(Closed):
		return Closed, nil
	default:
		return "", errors.New("invalid tender status")
	}
}
