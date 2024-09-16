package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	my_errors "tender-service/internal/errors"
	"tender-service/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type MockBidService struct{}

func (m *MockBidService) GetBidsByTenderID(tenderID, username string, limit, offset int) ([]models.Bid, error) {
	if tenderID == "invalid-uuid-format" {
		return nil, my_errors.ErrBadRequest
	}
	if tenderID == "non-existent-tender-id" {
		return nil, my_errors.ErrTenderNotFound
	}
	return []models.Bid{
		{
			ID:             "550e8400-e29b-41d4-a716-446655440099",
			TenderID:       tenderID,
			OrganizationID: "550e8400-e29b-41d4-a716-446655440022",
			UserID:         username,
			Description:    "Test bid",
			Status:         models.BidStatusCreated,
			AuthorType:     models.BidAuthorTypeUser,
		},
	}, nil
}

func (m *MockBidService) CreateBid(description, tenderID, organizationID, userID string, authorType models.BidAuthorType) (*models.Bid, error) {
	if tenderID == "invalid-uuid-format" || tenderID == "non-existent-tender-id" || userID == "non-existent-user-id" {
		return nil, my_errors.ErrBadRequest
	}

	if userID == "550e8400-e29b-41d4-a716-446655440008" {
		return nil, my_errors.ErrForbidden
	}
	return &models.Bid{
		ID:             "550e8400-e29b-41d4-a716-446655440099",
		TenderID:       tenderID,
		OrganizationID: organizationID,
		UserID:         userID,
		Description:    description,
		Status:         models.BidStatusCreated,
		AuthorType:     authorType,
	}, nil
}

func (m *MockBidService) GetUserBids(userID string, limit, offset int) ([]models.Bid, error) {
	if userID == "non-existent-user-id" {
		return nil, my_errors.ErrUserNotFound
	}
	return []models.Bid{
		{
			ID:             "550e8400-e29b-41d4-a716-446655440099",
			TenderID:       "446a0a79-ffdc-47ea-a91c-873f834c12a2",
			OrganizationID: "550e8400-e29b-41d4-a716-446655440022",
			UserID:         userID,
			Description:    "Test bid",
			Status:         models.BidStatusCreated,
			AuthorType:     models.BidAuthorTypeUser,
		},
	}, nil
}

func (m *MockBidService) GetBidStatus(bidID, username string) (models.BidStatus, error) {
	if bidID == "invalid-uuid-format" {
		return "", my_errors.ErrBadRequest
	}

	if bidID == "non-existent-bid-id" {
		return "", my_errors.ErrBidNotFound
	}

	if username == "unauthorized-user" {
		return "", my_errors.ErrForbidden
	}

	if bidID == "550e8400-e29b-41d4-a716-446655440099" && username == "user1" {
		return models.BidStatusCreated, nil
	}

	return "", my_errors.ErrBidNotFound
}

func (m *MockBidService) UpdateBidStatus(bidID, status, username string) error {
	_, err := uuid.Parse(bidID)
	if err != nil {
		return my_errors.ErrBadRequest
	}

	if bidID == "non-existent-bid-id" {
		return my_errors.ErrBidNotFound
	}

	if username == "unauthorized-user" {
		return my_errors.ErrForbidden
	}

	return nil
}

func (m *MockBidService) EditBid(bidID, username string, updatedFields map[string]interface{}) error {
	_, err := uuid.Parse(bidID)
	if err != nil {
		return my_errors.ErrBadRequest
	}

	if bidID == "non-existent-bid-id" {
		return my_errors.ErrBidNotFound
	}

	if username == "unauthorized-user" {
		return my_errors.ErrForbidden
	}

	return nil
}

func (m *MockBidService) SubmitBidFeedback(bidID, username, feedback string) (*models.Bid, error) {

	if bidID == "invalid-bid-id" {
		return nil, my_errors.ErrBadRequest
	}

	if bidID == "non-existent-bid-id" {
		return nil, my_errors.ErrBidNotFound
	}

	if username == "non-existent-user" {
		return nil, my_errors.ErrUserNotFound
	}

	if username == "unauthorized-user" {
		return nil, my_errors.ErrForbidden
	}

	bid := &models.Bid{
		ID:          bidID,
		Description: "Test Bid with feedback",
		Status:      models.BidStatusCreated,
	}
	return bid, nil
}

func TestCreateBid_Success(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	bid := models.Bid{
		Description:    "Успешное создание bid",
		TenderID:       "446a0a79-ffdc-47ea-a91c-873f834c12a2",
		OrganizationID: "550e8400-e29b-41d4-a716-446655440022",
		UserID:         "550e8400-e29b-41d4-a716-446655440003",
		AuthorType:     models.BidAuthorTypeUser,
	}
	requestBody := map[string]interface{}{
		"description":    bid.Description,
		"tenderId":       bid.TenderID,
		"organizationId": bid.OrganizationID,
		"userId":         bid.UserID,
		"authorType":     "User",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/api/bids/new", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateBid(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var createdBid models.Bid
	err = json.NewDecoder(rr.Body).Decode(&createdBid)
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440099", createdBid.ID)
}

func TestCreateBid_InvalidUUID(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	requestBody := map[string]interface{}{
		"description":    "Неверный формат tenderId",
		"tenderId":       "invalid-uuid-format",
		"organizationId": "550e8400-e29b-41d4-a716-446655440020",
		"userId":         "550e8400-e29b-41d4-a716-446655440001",
		"authorType":     "User",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/api/bids/new", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateBid(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBid_TenderNotFound(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	requestBody := map[string]interface{}{
		"description":    "Тест с несуществующим tenderId",
		"tenderId":       "non-existent-tender-id",
		"organizationId": "550e8400-e29b-41d4-a716-446655440020",
		"userId":         "550e8400-e29b-41d4-a716-446655440001",
		"authorType":     "User",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/api/bids/new", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateBid(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBid_UserNotFound(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	requestBody := map[string]interface{}{
		"description":    "Тест с несуществующим userId",
		"tenderId":       "446a0a79-ffdc-47ea-a91c-873f834c12a2",
		"organizationId": "550e8400-e29b-41d4-a716-446655440022",
		"userId":         "non-existent-user-id",
		"authorType":     "User",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/api/bids/new", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateBid(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBid_Forbidden(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	requestBody := map[string]interface{}{
		"description":    "Недостаточно прав для выполнения действия",
		"tenderId":       "446a0a79-ffdc-47ea-a91c-873f834c12a2",
		"organizationId": "550e8400-e29b-41d4-a716-446655440020",
		"userId":         "550e8400-e29b-41d4-a716-446655440008",
		"authorType":     "User",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/api/bids/new", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateBid(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestGetUserBids_Success(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/my?username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetUserBids(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var bids []models.Bid
	err = json.NewDecoder(rr.Body).Decode(&bids)
	assert.NoError(t, err)
	assert.Len(t, bids, 1)
	assert.Equal(t, "Test bid", bids[0].Description)
}

func TestGetUserBids_UserNotFound(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/my?username=non-existent-user-id", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetUserBids(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "User not found", errorResponse["reason"])
}

func TestGetBidsByTenderID_Success(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/446a0a79-ffdc-47ea-a91c-873f834c12a2/list?username=user1&limit=10&offset=0", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{tenderId}/list", handler.GetBidsByTenderID).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var bids []models.Bid
	err = json.NewDecoder(rr.Body).Decode(&bids)
	assert.NoError(t, err)
	assert.Len(t, bids, 1)
	assert.Equal(t, "Test bid", bids[0].Description)
}

func TestGetBidsByTenderID_TenderNotFound(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/non-existent-tender-id/list?username=user1&limit=10&offset=0", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{tenderId}/list", handler.GetBidsByTenderID).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid tender ID format", errorResponse["reason"])
}

func TestGetBidsByTenderID_InvalidOffset(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/446a0a79-ffdc-47ea-a91c-873f834c12a2/list?username=user1&limit=10&offset=-1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{tenderId}/list", handler.GetBidsByTenderID).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request parameters", errorResponse["reason"])
}

func TestGetBidsByTenderID_ZeroLimit(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/446a0a79-ffdc-47ea-a91c-873f834c12a2/list?username=user1&limit=0&offset=0", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{tenderId}/list", handler.GetBidsByTenderID).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request parameters", errorResponse["reason"])
}

func TestGetBidStatus_Success(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/550e8400-e29b-41d4-a716-446655440099/status?username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/status", handler.GetBidStatus).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	expectedStatus := "CREATED"
	assert.Equal(t, expectedStatus, rr.Body.String())
}

func TestGetBidStatus_BidNotFound(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/550e8400-e29b-41d4-a716-4466554400ff/status?username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/status", handler.GetBidStatus).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Bid not found", errorResponse["reason"])
}

func TestGetBidStatus_Forbidden(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("GET", "/api/bids/550e8400-e29b-41d4-a716-446655440008/status?username=unauthorized-user", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/status", handler.GetBidStatus).Methods("GET")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Insufficient permissions", errorResponse["reason"])
}

func TestUpdateBidStatus_Success(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("PUT", "/api/bids/550e8400-e29b-41d4-a716-446655440099/status?status=PUBLISHED&username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/status", handler.UpdateBidStatus).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateBidStatus_BidNotFound(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("PUT", "/api/bids/non-existent-bid-id/status?status=PUBLISHED&username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/status", handler.UpdateBidStatus).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid bid ID format", errorResponse["reason"])
}

func TestUpdateBidStatus_Forbidden(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("PUT", "/api/bids/550e8400-e29b-41d4-a716-446655440008/status?status=PUBLISHED&username=unauthorized-user", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/status", handler.UpdateBidStatus).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Insufficient permissions", errorResponse["reason"])
}

func TestEditBid_Success(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	requestBody := map[string]interface{}{
		"description": "Updated description for this bid",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("PATCH", "/api/bids/550e8400-e29b-41d4-a716-446655440099/edit?username=user1", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/edit", handler.EditBid).Methods("PATCH")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestEditBid_InvalidBidIDFormat(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	requestBody := map[string]interface{}{
		"description": "Updated description for this bid",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("PATCH", "/api/bids/invalid-bid-id/edit?username=user1", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/edit", handler.EditBid).Methods("PATCH")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid bid ID format", errorResponse["reason"])
}

func TestEditBid_Forbidden(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	requestBody := map[string]interface{}{
		"description": "Updated description for this bid",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("PATCH", "/api/bids/550e8400-e29b-41d4-a716-446655440008/edit?username=unauthorized-user", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/edit", handler.EditBid).Methods("PATCH")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var errorResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Insufficient permissions", errorResponse["reason"])
}

func TestSubmitBidFeedback_Success(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("PUT", "/api/bids/550e8400-e29b-41d4-a716-446655440099/feedback?username=user1&bidFeedback=Great%20job", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/feedback", handler.SubmitBidFeedback).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var bid models.Bid
	err = json.NewDecoder(rr.Body).Decode(&bid)
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440099", bid.ID)
}

func TestSubmitBidFeedback_BidNotFound(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("PUT", "/api/bids/non-existent-bid-id/feedback?username=user1&bidFeedback=Great%20job", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/feedback", handler.SubmitBidFeedback).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid bid ID format", response["reason"])
}

func TestSubmitBidFeedback_NoFeedback(t *testing.T) {
	mockService := &MockBidService{}
	handler := NewBidHandler(mockService)

	req, err := http.NewRequest("PUT", "/api/bids/550e8400-e29b-41d4-a716-446655440099/feedback?username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/bids/{bidId}/feedback", handler.SubmitBidFeedback).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Feedback is required", response["reason"])
}
