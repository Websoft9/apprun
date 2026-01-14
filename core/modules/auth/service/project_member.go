package service

import (
	"context"

	"apprun/ent"
	"apprun/internal/rbac"
	"apprun/modules/auth/repository"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
)

// ProjectMemberService handles project member business logic
type ProjectMemberService struct {
	memberRepo  *repository.ProjectMemberRepository
	projectRepo *repository.ProjectRepository
}

// NewProjectMemberService creates a new project member service
func NewProjectMemberService(
	memberRepo *repository.ProjectMemberRepository,
	projectRepo *repository.ProjectRepository,
) *ProjectMemberService {
	return &ProjectMemberService{
		memberRepo:  memberRepo,
		projectRepo: projectRepo,
	}
}

// AddMember adds a user to a project with a role
// This also syncs the role to Casbin
func (s *ProjectMemberService) AddMember(ctx context.Context, projectID, userID int64, role string) (*ent.ProjectMember, error) {
	// Validate role
	if !isValidProjectRole(role) {
		return nil, errors.Newf(errors.ErrCodeInvalidParam, "Invalid role: %s", role)
	}

	// Check if project exists
	exists, err := s.projectRepo.Exists(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to check project existence")
	}
	if !exists {
		return nil, errors.New(errors.ErrCodeNotFound, "Project not found")
	}

	// Add member to database
	member, err := s.memberRepo.AddMember(ctx, projectID, userID, role)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to add member")
	}

	// Add role to Casbin (domain-based RBAC)
	enforcer := rbac.GetEnforcer()
	domain := rbac.FormatDomain(projectID)
	_, err = enforcer.AddGroupingPolicy(rbac.FormatUserKey(userID), role, domain)
	if err != nil {
		// Rollback: remove member from database
		if rollbackErr := s.memberRepo.RemoveMember(ctx, member.ID); rollbackErr != nil {
			logger.Warn("Rollback failed when removing member",
				logger.Field{Key: "error", Value: rollbackErr},
				logger.Field{Key: "member_id", Value: member.ID})
		}
		return nil, errors.Wrap(err, errors.ErrCodeAuthPermCheckError, "Failed to add role to RBAC")
	}

	// Save Casbin policies
	if err := enforcer.SavePolicy(); err != nil {
		// Rollback
		if rollbackErr := s.memberRepo.RemoveMember(ctx, member.ID); rollbackErr != nil {
			logger.Warn("Rollback failed when removing member",
				logger.Field{Key: "error", Value: rollbackErr},
				logger.Field{Key: "member_id", Value: member.ID})
		}
		if _, rollbackErr := enforcer.RemoveGroupingPolicy(rbac.FormatUserKey(userID), role, domain); rollbackErr != nil {
			logger.Warn("Rollback failed when removing role",
				logger.Field{Key: "error", Value: rollbackErr},
				logger.Field{Key: "user_id", Value: userID},
				logger.Field{Key: "role", Value: role})
		}
		return nil, errors.Wrap(err, errors.ErrCodeAuthPermCheckError, "Failed to save RBAC policies")
	}

	return member, nil
}

// UpdateMemberRole updates a member's role in a project
func (s *ProjectMemberService) UpdateMemberRole(ctx context.Context, memberID int64, newRole string) (*ent.ProjectMember, error) {
	// Validate role
	if !isValidProjectRole(newRole) {
		return nil, errors.Newf(errors.ErrCodeInvalidParam, "Invalid role: %s", newRole)
	}

	// Get existing member
	member, err := s.memberRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeNotFound, "Member not found")
	}

	oldRole := member.Role

	// Update role in database
	member, err = s.memberRepo.UpdateRole(ctx, memberID, newRole)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to update role")
	}

	// Update role in Casbin
	enforcer := rbac.GetEnforcer()
	domain := rbac.FormatDomain(member.ProjectID)

	// Remove old role
	_, err = enforcer.RemoveGroupingPolicy(rbac.FormatUserKey(member.UserID), oldRole, domain)
	if err != nil {
		// Rollback
		if _, rollbackErr := s.memberRepo.UpdateRole(ctx, memberID, oldRole); rollbackErr != nil {
			logger.Warn("Rollback failed when updating role",
				logger.Field{Key: "error", Value: rollbackErr},
				logger.Field{Key: "member_id", Value: memberID})
		}
		return nil, errors.Wrap(err, errors.ErrCodeAuthPermCheckError, "Failed to remove old role from RBAC")
	}

	// Add new role
	_, err = enforcer.AddGroupingPolicy(rbac.FormatUserKey(member.UserID), newRole, domain)
	if err != nil {
		// Rollback
		if _, rollbackErr := s.memberRepo.UpdateRole(ctx, memberID, oldRole); rollbackErr != nil {
			logger.Warn("Rollback failed when updating role",
				logger.Field{Key: "error", Value: rollbackErr},
				logger.Field{Key: "member_id", Value: memberID})
		}
		if _, rollbackErr := enforcer.AddGroupingPolicy(rbac.FormatUserKey(member.UserID), oldRole, domain); rollbackErr != nil {
			logger.Warn("Rollback failed when adding old role",
				logger.Field{Key: "error", Value: rollbackErr},
				logger.Field{Key: "user_id", Value: member.UserID},
				logger.Field{Key: "role", Value: oldRole})
		}
		return nil, errors.Wrap(err, errors.ErrCodeAuthPermCheckError, "Failed to add new role to RBAC")
	}

	// Save policies and clear cache
	if err := enforcer.SavePolicy(); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeAuthPermCheckError, "Failed to save RBAC policies")
	}
	rbac.ClearUserCache(member.UserID, member.ProjectID)

	return member, nil
}

// RemoveMember removes a user from a project
func (s *ProjectMemberService) RemoveMember(ctx context.Context, memberID int64) error {
	// Get member details
	member, err := s.memberRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeNotFound, "Member not found")
	}

	// Remove from database
	if err := s.memberRepo.RemoveMember(ctx, memberID); err != nil {
		return errors.Wrap(err, errors.ErrCodeInternalError, "Failed to remove member")
	}

	// Remove role from Casbin
	enforcer := rbac.GetEnforcer()
	domain := rbac.FormatDomain(member.ProjectID)
	_, err = enforcer.RemoveGroupingPolicy(rbac.FormatUserKey(member.UserID), member.Role, domain)
	if err != nil {
		// Log error but don't fail - member already removed from DB
		logger.Warn("Failed to remove role from RBAC",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "user_id", Value: member.UserID},
			logger.Field{Key: "role", Value: member.Role})
	}

	// Save policies and clear cache
	if err := enforcer.SavePolicy(); err != nil {
		logger.Warn("Failed to save RBAC policies after member removal",
			logger.Field{Key: "error", Value: err})
	}
	rbac.ClearUserCache(member.UserID, member.ProjectID)

	return nil
}

// ListMembers retrieves all members of a project
func (s *ProjectMemberService) ListMembers(ctx context.Context, projectID int64) ([]*ent.ProjectMember, error) {
	return s.memberRepo.ListMembers(ctx, projectID)
}

// GetMemberRole retrieves a user's role in a project
func (s *ProjectMemberService) GetMemberRole(ctx context.Context, projectID, userID int64) (string, error) {
	return s.memberRepo.GetMemberRole(ctx, projectID, userID)
}

// IsMember checks if a user is a member of a project
func (s *ProjectMemberService) IsMember(ctx context.Context, projectID, userID int64) (bool, error) {
	return s.memberRepo.IsMember(ctx, projectID, userID)
}

// GetMemberByID retrieves a member by ID
func (s *ProjectMemberService) GetMemberByID(ctx context.Context, memberID int64) (*ent.ProjectMember, error) {
	return s.memberRepo.GetMemberByID(ctx, memberID)
}

// isValidProjectRole validates if a role is valid for projects
func isValidProjectRole(role string) bool {
	validRoles := map[string]bool{
		rbac.RoleProjectOwner:  true,
		rbac.RoleProjectAdmin:  true,
		rbac.RoleProjectMember: true,
		rbac.RoleProjectViewer: true,
	}
	return validRoles[role]
}
