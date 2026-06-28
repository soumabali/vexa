import { apiRequest } from "@/lib/api/client";

export type TeamRole = "owner" | "admin" | "member" | "viewer";

export interface Team {
  id: string;
  name: string;
  description?: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export interface TeamMember {
  id: string;
  team_id: string;
  user_id: string;
  role: TeamRole;
  joined_at: string;
  added_by?: string;
}

export interface CreateTeamInput {
  name: string;
  description?: string;
}

export interface UpdateTeamInput {
  name?: string;
  description?: string;
}

export interface AddMemberInput {
  user_id: string;
  role: TeamRole;
}

export interface UpdateMemberRoleInput {
  role: TeamRole;
}

export interface TeamsListResponse {
  teams: Team[];
}

export interface TeamResponse {
  team: Team;
}

export interface MembersListResponse {
  members: TeamMember[];
}

export interface MemberResponse {
  member: TeamMember;
}

export const teamsApi = {
  // List teams the current user belongs to
  list: (): Promise<Team[]> =>
    apiRequest<TeamsListResponse>("/api/v1/teams", { method: "GET" }).then(
      (d) => d.teams ?? []
    ),

  // Get a single team
  get: (id: string): Promise<Team> =>
    apiRequest<TeamResponse>(`/api/v1/teams/${id}`).then((d) => d.team),

  // Create a new team (caller becomes owner)
  create: (input: CreateTeamInput): Promise<Team> =>
    apiRequest<TeamResponse>("/api/v1/teams", {
      method: "POST",
      body: JSON.stringify(input),
    }).then((d) => d.team),

  // Update team name/description (owner only)
  update: (id: string, input: UpdateTeamInput): Promise<Team> =>
    apiRequest<TeamResponse>(`/api/v1/teams/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    }).then((d) => d.team),

  // Delete team (owner only)
  delete: (id: string): Promise<void> =>
    apiRequest<{ deleted: boolean }>(`/api/v1/teams/${id}`, {
      method: "DELETE",
    }).then(() => undefined),

  // List members of a team
  listMembers: (teamId: string): Promise<TeamMember[]> =>
    apiRequest<MembersListResponse>(
      `/api/v1/teams/${teamId}/members`,
      { method: "GET" }
    ).then((d) => d.members ?? []),

  // Add a member (owner/admin)
  addMember: (
    teamId: string,
    input: AddMemberInput
  ): Promise<TeamMember> =>
    apiRequest<MemberResponse>(`/api/v1/teams/${teamId}/members`, {
      method: "POST",
      body: JSON.stringify(input),
    }).then((d) => d.member),

  // Update member role (owner/admin)
  updateMemberRole: (
    teamId: string,
    userId: string,
    input: UpdateMemberRoleInput
  ): Promise<void> =>
    apiRequest<{ updated: boolean }>(
      `/api/v1/teams/${teamId}/members/${userId}`,
      { method: "PATCH", body: JSON.stringify(input) }
    ).then(() => undefined),

  // Remove a member (owner/admin)
  removeMember: (teamId: string, userId: string): Promise<void> =>
    apiRequest<{ removed: boolean }>(
      `/api/v1/teams/${teamId}/members/${userId}`,
      { method: "DELETE" }
    ).then(() => undefined),
};
