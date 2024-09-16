package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	my_errors "tender-service/internal/errors"
	"tender-service/internal/models"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type MockUserService struct{}

func (m *MockUserService) GetUserIDByUsername(username string) (string, error) {
	return "550e8400-e29b-41d4-a716-446655440002", nil
}

func (m *MockUserService) GetUserByUsername(username string) (*models.User, error) {
	if username == "non-existent-user" {
		return nil, my_errors.ErrUserNotFound
	}
	return &models.User{ID: "550e8400-e29b-41d4-a716-446655440002", Username: username}, nil
}

func (m *MockUserService) CheckUserPermission(userID, organizationID string) (bool, error) {
	if userID == "550e8400-e29b-41d4-a716-446655440002" && organizationID == "550e8400-e29b-41d4-a716-446655440020" {
		return true, nil
	}
	return false, my_errors.ErrForbidden
}

type MockTenderService struct{}

func (m *MockTenderService) GetTenders(serviceType string) ([]models.Tender, error) {
	return []models.Tender{
		{ID: "1", Name: "Test Tender", ServiceType: "Construction"},
	}, nil
}

func (m *MockTenderService) CreateTender(tender models.Tender, creatorUsername string) (models.Tender, error) {
	tender.ID = "21873f49-5776-4fb1-8866-aae300a08e45"
	return tender, nil
}

func (m *MockTenderService) GetUserTenders(username string) ([]models.Tender, error) {
	return []models.Tender{
		{ID: "1", Name: "User Tender", CreatorID: "550e8400-e29b-41d4-a716-446655440002"},
	}, nil
}

func (m *MockTenderService) GetTenderStatus(tenderId, username string) (models.TenderStatus, error) {
	if tenderId == "invalid-id" {
		return "", my_errors.ErrBadRequest
	}
	if tenderId == "nonexistent-tender-id" {
		return "", my_errors.ErrTenderNotFound
	}
	return models.Created, nil
}

func (m *MockTenderService) EditTender(tenderId, username string, name, description, serviceType *string) error {
	if tenderId == "invalid-id" {
		return my_errors.ErrBadRequest
	}
	if tenderId == "nonexistent-tender-id" {
		return my_errors.ErrTenderNotFound
	}
	return nil
}

func (m *MockTenderService) UpdateTenderStatus(tenderId string, status models.TenderStatus, username string) error {
	if tenderId == "invalid-id" {
		return my_errors.ErrBadRequest
	}
	if tenderId == "nonexistent-tender-id" {
		return my_errors.ErrTenderNotFound
	}
	return nil
}

func (m *MockTenderService) RollbackTenderVersion(tenderId string, version int, username string) error {
	if version == 999 {
		return my_errors.ErrTenderHistoryNotFound
	}
	if tenderId == "invalid-id" {
		return my_errors.ErrBadRequest
	}
	return nil
}

func TestGetTenders(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	req, err := http.NewRequest("GET", "/api/tenders", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetTenders(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var tenders []models.Tender
	err = json.NewDecoder(rr.Body).Decode(&tenders)
	assert.NoError(t, err)
	assert.Len(t, tenders, 1)
	assert.Equal(t, "Test Tender", tenders[0].Name)
}

func TestCreateTender(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	tender := models.Tender{
		Name:           "New Tender",
		Description:    "Tender Description",
		ServiceType:    "Construction",
		OrganizationID: "550e8400-e29b-41d4-a716-446655440020",
		Status:         models.Created,
	}
	requestBody := map[string]interface{}{
		"name":            tender.Name,
		"description":     tender.Description,
		"serviceType":     tender.ServiceType,
		"status":          "CREATED",
		"organizationId":  tender.OrganizationID,
		"creatorUsername": "user1",
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "/api/tenders/new", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTender(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var createdTender models.Tender
	err = json.NewDecoder(rr.Body).Decode(&createdTender)
	assert.NoError(t, err)
	assert.Equal(t, "21873f49-5776-4fb1-8866-aae300a08e45", createdTender.ID)
}

func TestUpdateTenderStatus_InvalidTenderID(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	req, err := http.NewRequest("PUT", "/api/tenders/invalid-id/status?status=PUBLISHED&username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/tenders/{tenderId}/status", handler.UpdateTenderStatus).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestEditTender_TenderNotFound(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	reqBody := `{"name": "Updated Tender Name"}`
	req, err := http.NewRequest("PATCH", "/api/tenders/nonexistent-tender-id/edit?username=user1", bytes.NewBuffer([]byte(reqBody)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/tenders/{tenderId}/edit", handler.EditTender).Methods("PATCH")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateTenderStatus_Success(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	req, err := http.NewRequest("PUT", "/api/tenders/d3bab548-a6bf-4838-9127-b40f77ec7812/status?status=PUBLISHED&username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/tenders/{tenderId}/status", handler.UpdateTenderStatus).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestEditTender_Success(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	reqBody := `{"name": "Updated Tender Name", "description": "Updated description"}`
	req, err := http.NewRequest("PATCH", "/api/tenders/d3bab548-a6bf-4838-9127-b40f77ec7812/edit?username=user1", bytes.NewBuffer([]byte(reqBody)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/tenders/{tenderId}/edit", handler.EditTender).Methods("PATCH")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRollbackTenderVersion_Success(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	req, err := http.NewRequest("PUT", "/api/tenders/d3bab548-a6bf-4838-9127-b40f77ec7812/rollback/2?username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", handler.RollbackTenderVersion).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRollbackTenderVersion_InvalidTenderID(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	req, err := http.NewRequest("PUT", "/api/tenders/invalid-id/rollback/2?username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", handler.RollbackTenderVersion).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRollbackTenderVersion_NonexistentVersion(t *testing.T) {
	mockService := &MockTenderService{}
	mockUserService := &MockUserService{}
	handler := NewTenderHandler(mockService, mockUserService)

	req, err := http.NewRequest("PUT", "/api/tenders/d3bab548-a6bf-4838-9127-b40f77ec7812/rollback/999?username=user1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", handler.RollbackTenderVersion).Methods("PUT")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
