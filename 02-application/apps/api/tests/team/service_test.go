package team_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/team"
)

func TestNewTeamService(t *testing.T) {
	ts := team.NewTeamService(nil, nil)
	require.NotNil(t, ts)
}

func TestTeamService_CreateTeam(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	ownerID := uuid.New()
	req := &team.CreateTeamRequest{
		Name:        "Test Team",
		Description: "A test team",
	}

	team, err := ts.CreateTeam(ownerID, req)
	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "Test Team", team.Name)
	assert.Equal(t, "A test team", team.Description)
	assert.Equal(t, ownerID, team.OwnerID)
	assert.True(t, team.IsActive)
	assert.NotEqual(t, uuid.Nil, team.ID)
}

func TestTeamService_GetTeam_NotFound(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	_, err := ts.GetTeam(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stub implementation")
}

func TestTeamService_ListTeams(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	teams, err := ts.ListTeams(uuid.New())
	require.NoError(t, err)
	assert.Empty(t, teams)
}

func TestTeamService_ListMembers(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	members, err := ts.ListMembers(uuid.New())
	require.NoError(t, err)
	assert.Empty(t, members)
}

func TestTeamService_ListSharedHosts(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	hosts, err := ts.ListSharedHosts(uuid.New())
	require.NoError(t, err)
	assert.Empty(t, hosts)
}

func TestTeamRole_Constants(t *testing.T) {
	assert.Equal(t, team.TeamRole("owner"), team.TeamRoleOwner)
	assert.Equal(t, team.TeamRole("admin"), team.TeamRoleAdmin)
	assert.Equal(t, team.TeamRole("member"), team.TeamRoleMember)
	assert.Equal(t, team.TeamRole("viewer"), team.TeamRoleViewer)
}

func TestTeamStructure(t *testing.T) {
	id := uuid.New()
	ownerID := uuid.New()
	team := &team.Team{
		ID:          id,
		Name:        "DevOps Team",
		Description: "The DevOps team",
		OwnerID:     ownerID,
		IsActive:    true,
	}

	assert.Equal(t, id, team.ID)
	assert.Equal(t, "DevOps Team", team.Name)
	assert.Equal(t, ownerID, team.OwnerID)
	assert.True(t, team.IsActive)
}

func TestTeamMemberStructure(t *testing.T) {
	member := &team.TeamMember{
		ID:     uuid.New(),
		TeamID: uuid.New(),
		UserID: uuid.New(),
		Role:   team.TeamRoleAdmin,
	}

	assert.Equal(t, team.TeamRoleAdmin, member.Role)
}

func TestSharedHostStructure(t *testing.T) {
	shared := &team.SharedHost{
		ID:         uuid.New(),
		TeamID:     uuid.New(),
		HostID:     uuid.New(),
		SharedBy:   uuid.New(),
		Permissions: []string{"connect", "edit"},
	}

	assert.Len(t, shared.Permissions, 2)
	assert.Contains(t, shared.Permissions, "connect")
	assert.Contains(t, shared.Permissions, "edit")
}

func TestCreateTeamRequest_Validation(t *testing.T) {
	req := &team.CreateTeamRequest{
		Name:        "My Team",
		Description: "Description here",
	}

	assert.Equal(t, "My Team", req.Name)
	assert.Equal(t, "Description here", req.Description)
}

func TestAddMemberRequest_Structure(t *testing.T) {
	userID := uuid.New()
	req := &team.AddMemberRequest{
		UserID: userID,
		Role:   team.TeamRoleMember,
	}

	assert.Equal(t, userID, req.UserID)
	assert.Equal(t, team.TeamRoleMember, req.Role)
}

func TestUpdateMemberRequest_Structure(t *testing.T) {
	req := &team.UpdateMemberRequest{
		Role: team.TeamRoleAdmin,
	}

	assert.Equal(t, team.TeamRoleAdmin, req.Role)
}

func TestShareHostRequest_Structure(t *testing.T) {
	hostID := uuid.New()
	req := &team.ShareHostRequest{
		HostID:      hostID,
		Permissions: []string{"connect"},
	}

	assert.Equal(t, hostID, req.HostID)
	assert.Len(t, req.Permissions, 1)
}

func TestTeamService_IsTeamAdmin_Stub(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	// Stub implementation returns false
	result := ts.IsTeamAdmin(uuid.New(), uuid.New())
	assert.False(t, result)
}

func TestTeamService_IsTeamMember_Stub(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	// Stub implementation returns false
	result := ts.IsTeamMember(uuid.New(), uuid.New())
	assert.False(t, result)
}

func TestTeamService_CanShare_Stub(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	// Stub implementation uses IsTeamAdmin which returns false
	result := ts.CanShare(uuid.New(), uuid.New())
	assert.False(t, result)
}

func TestTeamService_CanConnect_Stub(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	// Stub implementation returns false
	result := ts.CanConnect(uuid.New(), uuid.New(), uuid.New())
	assert.False(t, result)
}

func TestTeamService_CanEdit_Stub(t *testing.T) {
	ts := team.NewTeamService(nil, nil)

	// Stub implementation returns false
	result := ts.CanEdit(uuid.New(), uuid.New(), uuid.New())
	assert.False(t, result)
}

// BenchmarkCreateTeam benchmarks team creation
func BenchmarkCreateTeam(b *testing.B) {
	ts := team.NewTeamService(nil, nil)
	ownerID := uuid.New()
	req := &team.CreateTeamRequest{
		Name:        "Benchmark Team",
		Description: "Description",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ts.CreateTeam(ownerID, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListTeams benchmarks listing teams
func BenchmarkListTeams(b *testing.B) {
	ts := team.NewTeamService(nil, nil)
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.ListTeams(userID)
	}
}
