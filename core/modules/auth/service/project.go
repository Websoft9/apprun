// Package service provides business logic for authentication and authorization.
package service

import (
	"context"
	"fmt"

	"apprun/ent"
	"apprun/ent/schema"
	"apprun/internal/rbac"
	"apprun/modules/auth/repository"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
)

const (
	PlatformProjectName         = "Platform"
	PlatformProjectDescription  = "Global platform-level resources and configuration"
	PersonalProjectNameTemplate = "%s's Personal Project"
	PersonalProjectDescription  = "Personal workspace for individual user"
)

var (
	ErrProjectNotFound      = errors.New(errors.ErrCodeNotFound, "Project not found")
	ErrProjectAlreadyExists = errors.New(errors.ErrCodeConflict, "Project already exists")
)

type ProjectService struct {
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository, memberRepo *repository.ProjectMemberRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo, memberRepo: memberRepo}
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type ProjectResponse struct {
	UUID        string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int64  `json:"owner_id"`
	Status      int8   `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (s *ProjectService) CreateProject(ctx context.Context, name, description string, ownerID int64) (*ent.Project, error) {
	log := logger.L()
	project, err := s.projectRepo.Create(ctx, name, description, ownerID)
	if err != nil {
		log.Error("Failed to create project", logger.Field{Key: "error", Value: err.Error()}, logger.Field{Key: "name", Value: name})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to create project")
	}
	if _, err := s.memberRepo.AddMember(ctx, project.ID, ownerID, rbac.RoleProjectOwner); err != nil {
		log.Error("Failed to add owner", logger.Field{Key: "error", Value: err.Error()})
	}
	log.Info("Project created", logger.Field{Key: "project_id", Value: project.ID})
	return project, nil
}

func (s *ProjectService) CreatePersonalProject(ctx context.Context, userID int64, username string) (*ent.Project, error) {
	name := fmt.Sprintf(PersonalProjectNameTemplate, username)
	return s.CreateProject(ctx, name, PersonalProjectDescription, userID)
}

func (s *ProjectService) GetOrCreatePlatformProject(ctx context.Context, systemUserID int64) (*ent.Project, error) {
	log := logger.L()
	projects, err := s.projectRepo.ListByOwner(ctx, systemUserID)
	if err == nil {
		for _, p := range projects {
			if p.Name == PlatformProjectName {
				log.Info("Platform project exists", logger.Field{Key: "id", Value: p.ID})
				return p, nil
			}
		}
	}
	return s.CreateProject(ctx, PlatformProjectName, PlatformProjectDescription, systemUserID)
}

func (s *ProjectService) GetProjectByUUID(ctx context.Context, uuid string) (*ent.Project, error) {
	project, err := s.projectRepo.GetByUUID(ctx, uuid)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrProjectNotFound
		}
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to get project")
	}
	return project, nil
}

func (s *ProjectService) ListUserProjects(ctx context.Context, userID int64) ([]*ent.Project, error) {
	projects, err := s.projectRepo.ListByOwner(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to list projects")
	}
	return projects, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID int64, name, description *string) (*ent.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrProjectNotFound
		}
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to get project")
	}
	update := project.Update()
	if name != nil {
		update.SetName(*name)
	}
	if description != nil {
		update.SetDescription(*description)
	}
	return update.Save(ctx)
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID int64) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrProjectNotFound
		}
		return errors.Wrap(err, errors.ErrCodeInternalError, "Failed to get project")
	}
	_, err = project.Update().SetStatus(schema.ProjectStatusArchived).Save(ctx)
	return err
}

func ToProjectResponse(p *ent.Project) ProjectResponse {
	return ProjectResponse{
		UUID:        p.UUID,
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     p.OwnerID,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
