package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"tender-service/internal/models"
	"tender-service/internal/service"
	"tender-service/utils"

	my_errors "tender-service/internal/errors"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type BidHandler struct {
	bidService service.BidService
}

func NewBidHandler(bidService service.BidService) *BidHandler {
	return &BidHandler{bidService: bidService}
}

func (h *BidHandler) CreateBid(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Description    string `json:"description"`
		TenderID       string `json:"tenderId"`
		OrganizationID string `json:"organizationId"`
		UserID         string `json:"userId"`
		AuthorType     string `json:"authorType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdBid, err := h.bidService.CreateBid(
		request.Description,
		request.TenderID,
		request.OrganizationID,
		request.UserID,
		models.BidAuthorType(request.AuthorType),
	)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrInvalidUUID):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid UUID format")
		case errors.Is(err, my_errors.ErrTenderNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Tender not found")
		case errors.Is(err, my_errors.ErrUserNotFound):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case errors.Is(err, my_errors.ErrBadRequest):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error creating bid")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdBid)
}

func (h *BidHandler) GetUserBids(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		offset = 0
	}

	log.Printf("GetUserBids called for username: %s, limit: %d, offset: %d", username, limit, offset)

	bids, err := h.bidService.GetUserBids(username, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrUserNotFound):
			log.Printf("Error retrieving user by username: %s, error: %v", username, err)
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		default:
			log.Printf("Error retrieving bids for username: %s, error: %v", username, err)
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error retrieving bids")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bids)
}

func (h *BidHandler) GetBidsByTenderID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID := vars["tenderId"]

	if _, err := uuid.Parse(tenderID); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tender ID format")
		return
	}

	username := r.URL.Query().Get("username")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	bids, err := h.bidService.GetBidsByTenderID(tenderID, username, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrTenderNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Tender not found")
		case errors.Is(err, my_errors.ErrUserNotFound):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case errors.Is(err, my_errors.ErrBadRequest):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request parameters")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error retrieving bids")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bids)
}

func (h *BidHandler) GetBidStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidId"]

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	log.Printf("GetBidStatus: Received request for bidID=%s, username=%s", bidID, username)

	_, err := uuid.Parse(bidID)
	if err != nil {
		log.Printf("GetBidStatus: Invalid bidID format: %s", bidID)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid bid ID format")
		return
	}

	status, err := h.bidService.GetBidStatus(bidID, username)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrBidNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Bid not found")
		case errors.Is(err, my_errors.ErrUserNotFound):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		default:
			log.Printf("GetBidStatus: Internal server error for bidID=%s, username=%s: %v", bidID, username, err)
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	log.Printf("GetBidStatus: Successfully returned status for bidID=%s, username=%s", bidID, username)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(status))
}

func (h *BidHandler) UpdateBidStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidId"]

	status := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")
	if status == "" || username == "" {
		http.Error(w, "status and username are required", http.StatusBadRequest)
		return
	}

	log.Printf("UpdateBidStatus: Received request for bidID=%s, status=%s, username=%s", bidID, status, username)

	_, err := uuid.Parse(bidID)
	if err != nil {
		log.Printf("UpdateBidStatus: Invalid bidID format: %s", bidID)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid bid ID format")
		return
	}

	err = h.bidService.UpdateBidStatus(bidID, status, username)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrBidNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Bid not found")
		case errors.Is(err, my_errors.ErrUserNotFound):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		default:
			log.Printf("UpdateBidStatus: Internal server error for bidID=%s, status=%s, username=%s: %v", bidID, status, username, err)
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	log.Printf("UpdateBidStatus: Successfully updated status for bidID=%s, username=%s", bidID, username)
	w.WriteHeader(http.StatusOK)
}

func (h *BidHandler) EditBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidId"]

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	log.Printf("EditBid: Received request for bidID=%s, username=%s", bidID, username)

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("EditBid: Error decoding request body: %v", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.bidService.EditBid(bidID, username, payload)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrBidNotFound):
			utils.WriteErrorResponse(w, http.StatusNotFound, "Bid not found")
		case errors.Is(err, my_errors.ErrUserNotFound):
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case errors.Is(err, my_errors.ErrBadRequest):
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid bid ID format")
		default:
			log.Printf("EditBid: Internal server error for bidID=%s, username=%s: %v", bidID, username, err)
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	log.Printf("EditBid: Successfully edited bid for bidID=%s, username=%s", bidID, username)
	w.WriteHeader(http.StatusOK)
}

func (h *BidHandler) SubmitBidFeedback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidId"]

	username := r.URL.Query().Get("username")
	bidFeedback := r.URL.Query().Get("bidFeedback")

	if username == "" {
		log.Println("SubmitBidFeedback: Username is missing")
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Username is required")
		return
	}

	if bidFeedback == "" {
		log.Println("SubmitBidFeedback: Feedback is missing")
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Feedback is required")
		return
	}

	_, err := uuid.Parse(bidID)
	if err != nil {
		log.Printf("SubmitBidFeedback: Invalid bidID format: %s", bidID)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid bid ID format")
		return
	}

	log.Printf("SubmitBidFeedback: Received request for bidID=%s, username=%s, feedback=%s", bidID, username, bidFeedback)

	bid, err := h.bidService.SubmitBidFeedback(bidID, username, bidFeedback)
	if err != nil {
		switch {
		case errors.Is(err, my_errors.ErrBidNotFound):
			log.Printf("SubmitBidFeedback: Bid not found for bidID=%s", bidID)
			utils.WriteErrorResponse(w, http.StatusNotFound, "Bid not found")
		case errors.Is(err, my_errors.ErrUserNotFound):
			log.Printf("SubmitBidFeedback: User not found for username=%s", username)
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		case errors.Is(err, my_errors.ErrForbidden):
			log.Printf("SubmitBidFeedback: Insufficient permissions for username=%s on bidID=%s", username, bidID)
			utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		default:
			log.Printf("SubmitBidFeedback: Internal server error for bidID=%s, username=%s: %v", bidID, username, err)
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	log.Printf("SubmitBidFeedback: Successfully submitted feedback for bidID=%s, username=%s", bidID, username)

	response := map[string]interface{}{
		"id":         bid.ID,
		"name":       bid.Description,
		"status":     bid.Status,
		"authorType": bid.AuthorType,
		"authorId":   bid.UserID,
		// "version":    bid.Version,
		"createdAt": bid.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
