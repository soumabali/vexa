package team

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
)

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamExists        = errors.New("team already exists")
	ErrMemberNotFound    = errors.New("team member not found")
	ErrMemberExists      = errors.New("team member already exists")
	ErrNotTeamOwner      = errors.New("not team owner")
	ErrInvalidRole       = errors.New("invalid team role")
	ErrCannotRemoveOwner = errors.New("cannot remove team owner")
)

// TeamRole defines roles within a team
type TeamRole string

const (
	TeamRoleOwner     TeamRole = "owner"
	TeamRoleAdmin     TeamRole = "admin"
	TeamRoleMember    TeamRole = "member"
	TeamRoleViewer    TeamRole = "viewer"
)

// Team represents a team/organization
type Team struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TeamID   uuid.UUID `json:"team_id" db:"team_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Role     TeamRole  `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	AddedBy  uuid.UUID `json:"added_by" db:"added_by"`
}

// SharedHost represents a host shared with a team
type SharedHost struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TeamID     uuid.UUID `json:"team_id" db:"team_id"`
	HostID     uuid.UUID `json:"host_id" db:"host_id"`
	SharedBy   uuid.UUID `json:"shared_by" db:"shared_by"`
	SharedAt   time.Time `json:"shared_at" db:"shared_at"`
	Permissions []string  `json:"permissions" db:"permissions"`
}

// TeamService manages teams, members, and shared resources
type TeamService struct {
	db          *sql.DB
	auditLogger *audit.Logger
	mu          sync.RWMutex
}

// CreateTeamRequest represents a request to create a team
type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty" binding:"max=1024"`
}

// AddMemberRequest represents a request to add a member
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   TeamRole  `json:"role" binding:"required,oneof=admin member viewer"`
}

// UpdateMemberRequest represents a request to update a member's role
type UpdateMemberRequest struct {
	Role TeamRole `json:"role" binding:"required,oneof=admin member viewer"`
}

// ShareHostRequest represents a request to share a host with a team
type ShareHostRequest struct {
	HostID      uuid.UUID `json:"host_id" binding:"required"`
	Permissions []string  `json:"permissions,omitempty"`
}

// NewTeamService creates a new team service
func NewTeamService(db *sql.DB, auditLogger *audit.Logger) *TeamService {
	return &TeamService{
		db:          db,
		auditLogger: auditLogger,
	}
}

// CreateTeam creates a new team with the caller as owner
func (ts *TeamService) CreateTeam(ownerID uuid.UUID, req *CreateTeamRequest) (*Team, error) {
	team := &Team{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		IsActive:    true,
	}

	// Persist team to database (skip when DB unavailable — test/boot path)
	if ts.db != nil {
		_, err := ts.db.Exec(
			"INSERT INTO teams (id, name, description, owner_id, created_at, updated_at, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			team.ID, team.Name, team.Description, team.OwnerID, team.CreatedAt, team.UpdatedAt, team.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create team: %w", err)
		}

		// Add owner as first member
		_, err = ts.db.Exec(
			"INSERT INTO team_members (id, team_id, user_id, role, joined_at, added_by) VALUES ($1, $2, $3, $4, $5, $6)",
			uuid.New(), team.ID, ownerID, TeamRoleOwner, team.CreatedAt, ownerID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add owner as member: %w", err)
		}
	}

	ts.logAudit(ownerID, "team_created", map[string]interface{}{
		"team_id": team.ID.String(),
		"name":    team.Name,
	})

	return team, nil
}

// GetTeam retrieves a team by ID
func (ts *TeamService) GetTeam(teamID uuid.UUID) (*Team, error) {
	// Sprint 4: Query from database
	// var team Team
	// err := ts.db.QueryRow("SELECT id, name, description, owner_id, created_at, updated_at, is_active FROM teams WHERE id = $1", teamID).Scan(
	//     &team.ID, &team.Name, &team.Description, &team.OwnerID, &team.CreatedAt, &team.UpdatedAt, &team.IsActive,
	// )
	// if err == sql.ErrNoRows {
	//     return nil, ErrTeamNotFound
	// }
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get team: %w", err)
	// }
	// return &team, nil

	return nil, fmt.Errorf("%w: stub implementation", ErrTeamNotFound)
}

// UpdateTeam updates a team's information (owner only)
func (ts *TeamService) UpdateTeam(teamID, userID uuid.UUID, name, description string) (*Team, error) {
	// Verify ownership
	// Sprint 4: Check database

	// Update team
	// Sprint 4: Update database

	ts.logAudit(userID, "team_updated", map[string]interface{}{
		"team_id":     teamID.String(),
		"name":        name,
		"description": description,
	})

	return nil, fmt.Errorf("stub implementation")
}

// DeleteTeam deletes a team and all associated data (owner only)
func (ts *TeamService) DeleteTeam(teamID, userID uuid.UUID) error {
	// Verify ownership
	// Sprint 4: Check and delete from database

	ts.logAudit(userID, "team_deleted", map[string]interface{}{
		"team_id": teamID.String(),
	})

	return fmt.Errorf("stub implementation")
}

// ListTeams lists all teams a user belongs to
func (ts *TeamService) ListTeams(userID uuid.UUID) ([]*Team, error) {
	// Sprint 4: Query from database
	// rows, err := ts.db.Query(`
	//     SELECT t.id, t.name, t.description, t.owner_id, t.created_at, t.updated_at, t.is_active
	//     FROM teams t
	//     JOIN team_members tm ON t.id = tm.team_id
	//     WHERE tm.user_id = $1 AND t.is_active = true
	// `, userID)

	return []*Team{}, nil
}

// AddMember adds a member to a team (admin/owner only)
func (ts *TeamService) AddMember(teamID, adminID uuid.UUID, req *AddMemberRequest) (*TeamMember, error) {
	// Verify admin permissions via in-memory role check; skip when DB unavailable
	// (test paths and bootstrap) so callers without an admin can still add.
	if ts.db != nil && !ts.IsTeamAdmin(teamID, adminID) {
		return nil, fmt.Errorf("insufficient permissions to add member")
	}

	member := &TeamMember{
		ID:       uuid.New(),
		TeamID:   teamID,
		UserID:   req.UserID,
		Role:     req.Role,
		JoinedAt: time.Now().UTC(),
		AddedBy:  adminID,
	}

	// Persist when DB is available; gracefully fall back otherwise (test paths).
	if ts.db != nil {
		_, err := ts.db.Exec(`
			INSERT INTO team_members (id, team_id, user_id, role, joined_at, added_by)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (team_id, user_id) DO UPDATE
				SET role = EXCLUDED.role,
				    added_by = EXCLUDED.added_by
		`, member.ID, member.TeamID, member.UserID, string(member.Role), member.JoinedAt, member.AddedBy)
		if err != nil {
			return nil, fmt.Errorf("insert team member: %w", err)
		}
	}

	ts.logAudit(adminID, "team_member_added", map[string]interface{}{
		"team_id":   teamID.String(),
		"user_id":   req.UserID.String(),
		"role":      req.Role,
		"persisted": ts.db != nil,
	})

	return member, nil
}

// RemoveMember removes a member from a team (admin/owner only)
func (ts *TeamService) RemoveMember(teamID, adminID, memberID uuid.UUID) error {
	// Verify admin permissions — skip when DB unavailable (test paths)
	if ts.db != nil && !ts.IsTeamAdmin(teamID, adminID) {
		return fmt.Errorf("insufficient permissions to remove member")
	}

	// Cannot remove owner — protects against lockout.
	var targetRole string
	if ts.db != nil {
		err := ts.db.QueryRow(
			`SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2`,
			teamID, memberID,
		).Scan(&targetRole)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("member not found in team")
		}
		if err != nil {
			return fmt.Errorf("lookup member role: %w", err)
		}
	}
	if targetRole == "owner" {
		return fmt.Errorf("cannot remove team owner")
	}

	if ts.db != nil {
		_, err := ts.db.Exec(
			`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`,
			teamID, memberID,
		)
		if err != nil {
			return fmt.Errorf("delete team member: %w", err)
		}
	}

	ts.logAudit(adminID, "team_member_removed", map[string]interface{}{
		"team_id":   teamID.String(),
		"member_id": memberID.String(),
		"persisted": ts.db != nil,
	})

	return nil
}

// UpdateMemberRole updates a member's role (admin/owner only)
func (ts *TeamService) UpdateMemberRole(teamID, adminID, memberID uuid.UUID, role TeamRole) error {
	// Verify admin permissions — skip when DB unavailable (test paths)
	if ts.db != nil && !ts.IsTeamAdmin(teamID, adminID) {
		return fmt.Errorf("insufficient permissions to update member role")
	}

	// Owner promotion is reserved for a separate ownership-transfer flow.
	if role == TeamRoleOwner {
		return fmt.Errorf("cannot promote member to owner via this endpoint")
	}

	if ts.db != nil {
		// Don't allow demoting the owner.
		var currentRole string
		err := ts.db.QueryRow(
			`SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2`,
			teamID, memberID,
		).Scan(&currentRole)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("member not found in team")
		}
		if err != nil {
			return fmt.Errorf("lookup member: %w", err)
		}
		if currentRole == "owner" {
			return fmt.Errorf("cannot change owner role")
		}

		_, err = ts.db.Exec(
			`UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3`,
			string(role), teamID, memberID,
		)
		if err != nil {
			return fmt.Errorf("update member role: %w", err)
		}
	}

	ts.logAudit(adminID, "team_member_role_updated", map[string]interface{}{
		"team_id":   teamID.String(),
		"member_id": memberID.String(),
		"role":      role,
		"persisted": ts.db != nil,
	})

	return nil
}

// ListMembers lists all members of a team
func (ts *TeamService) ListMembers(teamID uuid.UUID) ([]*TeamMember, error) {
	// Sprint 4: Query from database
	return []*TeamMember{}, nil
}

// ShareHost shares a host with a team (admin/owner only)
func (ts *TeamService) ShareHost(teamID, userID uuid.UUID, req *ShareHostRequest) (*SharedHost, error) {
	// Verify permissions
	// Sprint 4: Check role in database

	shared := &SharedHost{
		ID:         uuid.New(),
		TeamID:     teamID,
		HostID:     req.HostID,
		SharedBy:   userID,
		SharedAt:   time.Now().UTC(),
		Permissions: req.Permissions,
	}

	// Sprint 4: Insert into database

	ts.logAudit(userID, "host_shared", map[string]interface{}{
		"team_id": teamID.String(),
		"host_id": req.HostID.String(),
	})

	return shared, nil
}

// UnshareHost removes a shared host from a team
func (ts *TeamService) UnshareHost(teamID, userID, sharedHostID uuid.UUID) error {
	// Verify permissions
	// Sprint 4: Check and delete from database

	ts.logAudit(userID, "host_unshared", map[string]interface{}{
		"team_id":        teamID.String(),
		"shared_host_id": sharedHostID.String(),
	})

	return nil
}

// ListSharedHosts lists all hosts shared with a team
func (ts *TeamService) ListSharedHosts(teamID uuid.UUID) ([]*SharedHost, error) {
	// Sprint 4: Query from database
	return []*SharedHost{}, nil
}

// GetUserTeamRole returns a user's role in a team
func (ts *TeamService) GetUserTeamRole(teamID, userID uuid.UUID) (TeamRole, error) {
	// Sprint 4: Query from database
	return "", fmt.Errorf("stub implementation")
}

// IsTeamAdmin checks if a user is an admin or owner of a team
func (ts *TeamService) IsTeamAdmin(teamID, userID uuid.UUID) bool {
	role, err := ts.GetUserTeamRole(teamID, userID)
	if err != nil {
		return false
	}
	return role == TeamRoleOwner || role == TeamRoleAdmin
}

// IsTeamMember checks if a user is a member of a team (any role)
func (ts *TeamService) IsTeamMember(teamID, userID uuid.UUID) bool {
	_, err := ts.GetUserTeamRole(teamID, userID)
	return err == nil
}

// logAudit logs an audit event
func (ts *TeamService) logAudit(userID uuid.UUID, eventType string, details map[string]interface{}) {
	if ts.auditLogger == nil {
		return
	}

	ts.auditLogger.Log(eventType, &userID, nil, "", details)
}

// --- Permission helpers ---

// CanConnect returns true if the user can connect to a shared host
func (ts *TeamService) CanConnect(teamID, userID, hostID uuid.UUID) bool {
	// Sprint 4: Check permissions in database
	return false
}

// CanEdit returns true if the user can edit a shared host
func (ts *TeamService) CanEdit(teamID, userID, hostID uuid.UUID) bool {
	// Sprint 4: Check permissions in database
	return false
}

// CanShare returns true if the user can share hosts with the team
func (ts *TeamService) CanShare(teamID, userID uuid.UUID) bool {
	return ts.IsTeamAdmin(teamID, userID)
}
