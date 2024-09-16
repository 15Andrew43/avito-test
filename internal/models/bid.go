package models

type BidAuthorType string

const (
	BidAuthorTypeUser         BidAuthorType = "User"
	BidAuthorTypeOrganization BidAuthorType = "Organization"
)

type Bid struct {
	ID             string
	TenderID       string
	OrganizationID string
	UserID         string
	AuthorType     BidAuthorType
	Description    string
	Status         BidStatus
	CreatedAt      string
	UpdatedAt      string
}

type BidStatus string

const (
	BidStatusCreated   BidStatus = "CREATED"
	BidStatusPublished BidStatus = "PUBLISHED"
	BidStatusCanceled  BidStatus = "CANCELED"
)
