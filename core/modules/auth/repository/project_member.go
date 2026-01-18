package repository

import (
	"context"

	"apprun/ent"
	"apprun/ent/projectmember"
)

// ProjectMemberRepository handles project member data access
type ProjectMemberRepository struct {
	client *ent.Client
}

// NewProjectMemberRepository creates a new project member repository
func NewProjectMemberRepository(client *ent.Client) *ProjectMemberRepository {
	return &ProjectMemberRepository{client: client}
}

// AddMember adds a user to a project with a role
func (r *ProjectMemberRepository) AddMember(ctx context.Context, projectID, userID int64, role string) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.
		Create().
		SetProjectID(projectID).
		SetUserID(userID).
		SetRole(role).
		Save(ctx)
}

// GetMember retrieves a specific project member
func (r *ProjectMemberRepository) GetMember(ctx context.Context, projectID, userID int64) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.
		Query().
		Where(
			projectmember.ProjectIDEQ(projectID),
			projectmember.UserIDEQ(userID),
		).
		Only(ctx)
}

// GetMemberByID retrieves a project member by ID
func (r *ProjectMemberRepository) GetMemberByID(ctx context.Context, id int64) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.
		Query().
		Where(projectmember.IDEQ(id)).
		Only(ctx)
}

// ListMembers retrieves all members of a project
func (r *ProjectMemberRepository) ListMembers(ctx context.Context, projectID int64) ([]*ent.ProjectMember, error) {
	return r.client.ProjectMember.
		Query().
		Where(projectmember.ProjectIDEQ(projectID)).
		WithUser(). // Load user information
		All(ctx)
}

// ListUserProjects retrieves all projects where a user is a member
func (r *ProjectMemberRepository) ListUserProjects(ctx context.Context, userID int64) ([]*ent.ProjectMember, error) {
	return r.client.ProjectMember.
		Query().
		Where(projectmember.UserIDEQ(userID)).
		WithProject(). // Load project information
		All(ctx)
}

// UpdateRole updates a member's role in a project
func (r *ProjectMemberRepository) UpdateRole(ctx context.Context, id int64, role string) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.
		UpdateOneID(id).
		SetRole(role).
		Save(ctx)
}

// RemoveMember removes a user from a project
func (r *ProjectMemberRepository) RemoveMember(ctx context.Context, id int64) error {
	return r.client.ProjectMember.DeleteOneID(id).Exec(ctx)
}

// RemoveMemberByUserID removes a user from a project by user ID
func (r *ProjectMemberRepository) RemoveMemberByUserID(ctx context.Context, projectID, userID int64) error {
	_, err := r.client.ProjectMember.
		Delete().
		Where(
			projectmember.ProjectIDEQ(projectID),
			projectmember.UserIDEQ(userID),
		).
		Exec(ctx)
	return err
}

// IsMember checks if a user is a member of a project
func (r *ProjectMemberRepository) IsMember(ctx context.Context, projectID, userID int64) (bool, error) {
	count, err := r.client.ProjectMember.
		Query().
		Where(
			projectmember.ProjectIDEQ(projectID),
			projectmember.UserIDEQ(userID),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMemberRole retrieves a user's role in a project
func (r *ProjectMemberRepository) GetMemberRole(ctx context.Context, projectID, userID int64) (string, error) {
	member, err := r.GetMember(ctx, projectID, userID)
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

// CountMembers counts the number of members in a project
func (r *ProjectMemberRepository) CountMembers(ctx context.Context, projectID int64) (int, error) {
	return r.client.ProjectMember.
		Query().
		Where(projectmember.ProjectIDEQ(projectID)).
		Count(ctx)
}
