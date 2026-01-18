package rbac

// Platform-level roles (global roles)
const (
	RolePlatformAdmin = "platform_admin" // Platform administrator
	RolePlatformUser  = "platform_user"  // Regular user
)

// Project-level roles
const (
	RoleProjectOwner  = "owner"  // Project owner
	RoleProjectAdmin  = "admin"  // Project administrator
	RoleProjectMember = "member" // Project member
	RoleProjectViewer = "viewer" // Viewer (read-only)
)

// PlatformDomain is the domain constant for platform-level resources
const PlatformDomain = "platform"

// Resources (resource types)
const (
	ResourceConfig   = "config"   // Configuration management
	ResourceData     = "data"     // Data model
	ResourceStorage  = "storage"  // File storage
	ResourceFunction = "function" // Function service
	ResourceWorkflow = "workflow" // Workflow
	ResourceMember   = "member"   // Member management
	ResourceProject  = "project"  // Project settings
)

// Actions (operation types)
const (
	ActionCreate  = "create"
	ActionRead    = "read"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionExecute = "execute"
	ActionManage  = "manage"
)
