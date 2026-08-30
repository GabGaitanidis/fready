package group

import (
	"encoding/json"
	"fready/internal/auth/session"
	"fready/internal/response"
	"log/slog"
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
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid body")
		return
	}

	g, err := h.service.CreateGroup(r.Context(), req.Name, userID)
	if err != nil {
		slog.Error("failed to create group", "error", err, "user_id", userID)
		response.Error(w, http.StatusBadRequest, "Failed to create group")
		return
	}

	response.Success(w, http.StatusCreated, g)
}

func (h *Handler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	groups, err := h.service.ListMyGroups(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list groups", "error", err, "user_id", userID)
		response.Error(w, http.StatusInternalServerError, "Failed to list groups")
		return
	}

	response.Success(w, http.StatusOK, groups)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	g, err := h.service.GetGroup(r.Context(), groupID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Group not found")
		return
	}

	response.Success(w, http.StatusOK, g)
}

type addMemberRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	requesterID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid body")
		return
	}

	if err := h.service.AddMember(r.Context(), groupID, requesterID, req.UserID); err != nil {
		slog.Warn("add member denied or failed", "error", err, "group_id", groupID, "requester_id", requesterID)
		response.Error(w, http.StatusForbidden, "Not authorized to add this member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	requesterID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	memberID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	if err := h.service.RemoveMember(r.Context(), groupID, requesterID, memberID); err != nil {
		slog.Warn("remove member denied or failed", "error", err, "group_id", groupID, "requester_id", requesterID)
		response.Error(w, http.StatusForbidden, "Not authorized to remove this member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	members, err := h.service.ListMembers(r.Context(), groupID)
	if err != nil {
		slog.Error("failed to list group members", "error", err, "group_id", groupID)
		response.Error(w, http.StatusInternalServerError, "Failed to list members")
		return
	}

	response.Success(w, http.StatusOK, members)
}