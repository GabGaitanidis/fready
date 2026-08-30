package group

import (
	"encoding/json"
	"fready/internal/auth/session"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service        Service
	sessionService session.Service
}



func NewHandler(s Service, ss session.Service) *Handler {
	return &Handler{service: s, sessionService: ss}
}

type createGroupRequest struct {
	Name string `json:"name"`
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	g, err := h.service.CreateGroup(r.Context(), req.Name, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func (h *Handler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	groups, err := h.service.ListMyGroups(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to list groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group id", http.StatusBadRequest)
		return
	}

	g, err := h.service.GetGroup(r.Context(), groupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

type addMemberRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	requesterID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group id", http.StatusBadRequest)
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := h.service.AddMember(r.Context(), groupID, requesterID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	requesterID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group id", http.StatusBadRequest)
		return
	}

	memberID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveMember(r.Context(), groupID, requesterID, memberID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group id", http.StatusBadRequest)
		return
	}

	members, err := h.service.ListMembers(r.Context(), groupID)
	if err != nil {
		http.Error(w, "Failed to list members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}