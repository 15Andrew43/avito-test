package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"tender-service/internal/models"
	"tender-service/internal/service"

	"github.com/gorilla/mux"

	"tender-service/utils"

	my_errors "tender-service/internal/errors"
)

type TenderHandler struct {
	tenderService service.TenderService
	userService   service.UserService
}

func NewTenderHandler(tenderService service.TenderService, userService service.UserService) *TenderHandler {
	return &TenderHandler{
		tenderService: tenderService,
		userService:   userService,
	}
}

func (h *TenderHandler) GetTenders(w http.ResponseWriter, r *http.Request) {
	serviceType := r.URL.Query().Get("service_type")

	tenders, err := h.tenderService.GetTenders(serviceType)
	if err != nil {
		log.Printf("Error fetching tenders: %v", err)
		http.Error(w, "Error fetching tenders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tenders); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *TenderHandler) GetUserTenders(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	tenders, err := h.tenderService.GetUserTenders(username)
	if err != nil {
		log.Printf("Error fetching user tenders: %v", err)
		http.Error(w, "Error fetching user tenders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tenders); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *TenderHandler) CreateTender(w http.ResponseWriter, r *http.Request) {

	var request struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		ServiceType     string `json:"serviceType"`
		Status          string `json:"status"`
		OrganizationID  string `json:"organizationId"`
		CreatorUsername string `json:"creatorUsername"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.CreatorUsername == "" || request.Status == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	tender := models.Tender{
		Name:           request.Name,
		Description:    request.Description,
		ServiceType:    request.ServiceType,
		Status:         models.TenderStatus(request.Status),
		OrganizationID: request.OrganizationID,
	}
	createdTender, err := h.tenderService.CreateTender(tender, request.CreatorUsername)
	if err != nil {
		if errors.Is(err, my_errors.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Creator user not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error creating tender")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdTender)
}

func (h *TenderHandler) GetTenderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]

	username := r.URL.Query().Get("username")
	if username == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Missing username")
		return
	}

	status, err := h.tenderService.GetTenderStatus(tenderId, username)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrTenderNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Tender not found")
		case errors.Is(err, my_errors.ErrUnauthorized):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case errors.Is(err, my_errors.ErrBadRequest):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tender ID")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error fetching tender status")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *TenderHandler) UpdateTenderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]

	statusStr := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")

	if statusStr == "" || username == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	status, err := models.ParseTenderStatus(statusStr)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tender status")
		return
	}

	err = h.tenderService.UpdateTenderStatus(tenderId, status, username)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrTenderNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Tender not found")
		case errors.Is(err, my_errors.ErrUnauthorized):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case errors.Is(err, my_errors.ErrBadRequest):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid bid ID format")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error updating tender status")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TenderHandler) EditTender(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]

	username := r.URL.Query().Get("username")
	if username == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Missing username")
		return
	}

	var request struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		ServiceType *string `json:"serviceType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.tenderService.EditTender(tenderId, username, request.Name, request.Description, request.ServiceType)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrTenderNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Tender not found")
		case errors.Is(err, my_errors.ErrUnauthorized):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case errors.Is(err, my_errors.ErrBadRequest):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request data")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error editing tender")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TenderHandler) RollbackTenderVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderId := vars["tenderId"]
	versionStr := vars["version"]
	username := r.URL.Query().Get("username")

	if username == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil || version <= 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid version number")
		return
	}

	err = h.tenderService.RollbackTenderVersion(tenderId, version, username)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrTenderNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Tender not found")
		case errors.Is(err, my_errors.ErrTenderHistoryNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Tender version not found")
		case errors.Is(err, my_errors.ErrUnauthorized):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case errors.Is(err, my_errors.ErrBadRequest):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tender ID or version")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error rolling back tender")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
