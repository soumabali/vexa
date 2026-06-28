package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/team"
)

type TeamHandler struct {
	teamService *team.TeamService
	auditLogger *audit.Logger
}

func NewTeamHandler(teamService *team.TeamService, auditLogger *audit.Logger) *TeamHandler {
	return &TeamHandler{teamService: teamService, auditLogger: auditLogger}
}

// getUserID extracts authenticated userID from gin context.
func getTeamUserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	uid, ok := v.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return uuid.Nil, false
	}
	return uid, true
}

// Create POST /api/v1/teams
func (h *TeamHandler) Create(c *gin.Context) {
	userID, ok := getTeamUserID(c)
	if !ok {
		return
	}

	var req team.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.teamService.CreateTeam(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log("team.create", &userID, nil, c.ClientIP(), map[string]interface{}{
			"team_id": t.ID.String(),
			"name":    t.Name,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"team": t})
}

// List GET /api/v1/teams
func (h *TeamHandler) List(c *gin.Context) {
	userID, ok := getTeamUserID(c)
	if !ok {
		return
	}

	teams, err := h.teamService.ListTeams(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teams"})
		return
	}
	if teams == nil {
		teams = []*team.Team{}
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

// Get GET /api/v1/teams/:id
func (h *TeamHandler) Get(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	t, err := h.teamService.GetTeam(teamID)
	if err != nil {
		if errors.Is(err, team.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": t})
}

// Update PATCH /api/v1/teams/:id
func (h *TeamHandler) Update(c *gin.Context) {
	userID, ok := getTeamUserID(c)
	if !ok {
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var req team.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.teamService.UpdateTeam(teamID, userID, req.Name, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		case errors.Is(err, team.ErrNotTeamOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "owner role required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update team"})
		}
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log("team.update", &userID, nil, c.ClientIP(), map[string]interface{}{
			"team_id": t.ID.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"team": t})
}

// Delete DELETE /api/v1/teams/:id
func (h *TeamHandler) Delete(c *gin.Context) {
	userID, ok := getTeamUserID(c)
	if !ok {
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	if err := h.teamService.DeleteTeam(teamID, userID); err != nil {
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		case errors.Is(err, team.ErrNotTeamOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "owner role required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete team"})
		}
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log("team.delete", &userID, nil, c.ClientIP(), map[string]interface{}{
			"team_id": teamID.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// AddMember POST /api/v1/teams/:id/members
func (h *TeamHandler) AddMember(c *gin.Context) {
	userID, ok := getTeamUserID(c)
	if !ok {
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var req team.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.teamService.AddMember(teamID, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		case errors.Is(err, team.ErrMemberExists):
			c.JSON(http.StatusConflict, gin.H{"error": "member already exists"})
		case errors.Is(err, team.ErrNotTeamOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "owner or admin role required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
		}
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log("team.member.add", &userID, nil, c.ClientIP(), map[string]interface{}{
			"team_id":  teamID.String(),
			"member":   member.UserID.String(),
			"role":     string(member.Role),
		})
	}

	c.JSON(http.StatusCreated, gin.H{"member": member})
}

// RemoveMember DELETE /api/v1/teams/:id/members/:user_id
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	userID, ok := getTeamUserID(c)
	if !ok {
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}
	memberID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.teamService.RemoveMember(teamID, userID, memberID); err != nil {
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		case errors.Is(err, team.ErrMemberNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		case errors.Is(err, team.ErrCannotRemoveOwner):
			c.JSON(http.StatusConflict, gin.H{"error": "cannot remove team owner"})
		case errors.Is(err, team.ErrNotTeamOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "owner or admin role required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
		}
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log("team.member.remove", &userID, nil, c.ClientIP(), map[string]interface{}{
			"team_id": teamID.String(),
			"member":  memberID.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"removed": true})
}

// UpdateMemberRole PATCH /api/v1/teams/:id/members/:user_id
func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	userID, ok := getTeamUserID(c)
	if !ok {
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}
	memberID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req team.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.teamService.UpdateMemberRole(teamID, userID, memberID, req.Role); err != nil {
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		case errors.Is(err, team.ErrMemberNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		case errors.Is(err, team.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		case errors.Is(err, team.ErrNotTeamOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "owner or admin role required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		}
		return
	}

	if h.auditLogger != nil {
		h.auditLogger.Log("team.member.update_role", &userID, nil, c.ClientIP(), map[string]interface{}{
			"team_id": teamID.String(),
			"member":  memberID.String(),
			"role":    string(req.Role),
		})
	}

	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// ListMembers GET /api/v1/teams/:id/members
func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	members, err := h.teamService.ListMembers(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list members"})
		return
	}
	if members == nil {
		members = []*team.TeamMember{}
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}
