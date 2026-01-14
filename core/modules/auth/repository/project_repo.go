package repository

import (
	"context"
	"fmt"

	"apprun/ent"
	"apprun/ent/project"
)

// ProjectRepository handles project data access
type ProjectRepository struct {
	client *ent.Client
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(client *ent.Client) *ProjectRepository {
	return &ProjectRepository{client: client}
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, name, description string, ownerID int64) (*ent.Project, error) {
	return r.client.Project.
		Create().
		SetName(name).
		SetDescription(description).
		SetOwnerID(ownerID).
		SetStatus(1). // 1 = enabled
		Save(ctx)
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*ent.Project, error) {
	return r.client.Project.
		Query().
		Where(project.IDEQ(id)).
		Only(ctx)
}

// GetByUUID retrieves a project by UUID
func (r *ProjectRepository) GetByUUID(ctx context.Context, uuid string) (*ent.Project, error) {
	return r.client.Project.
		Query().
		Where(project.UUIDEQ(uuid)).
		Only(ctx)
}

// ListByOwner retrieves all projects owned by a user
func (r *ProjectRepository) ListByOwner(ctx context.Context, ownerID int64) ([]*ent.Project, error) {
	return r.client.Project.
		Query().
		Where(project.OwnerIDEQ(ownerID)).
		All(ctx)
}

// Update updates a project
func (r *ProjectRepository) Update(ctx context.Context, id int64, name, description string) (*ent.Project, error) {
	return r.client.Project.
		UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		Save(ctx)
}

// UpdateStatus updates project status
func (r *ProjectRepository) UpdateStatus(ctx context.Context, id int64, status int8) error {
	return r.client.Project.
		UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
}

// Delete deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	return r.client.Project.DeleteOneID(id).Exec(ctx)
}

// Exists checks if a project exists
func (r *ProjectRepository) Exists(ctx context.Context, id int64) (bool, error) {
	count, err := r.client.Project.
		Query().
		Where(project.IDEQ(id)).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	return count > 0, nil
}
