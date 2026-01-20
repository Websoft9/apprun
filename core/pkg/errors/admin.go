package errors

// Predefined admin module errors
var (
	ErrAdminEmailExists           = New(ErrCodeAdminEmailExists, "Email already registered")
	ErrAdminCannotDeleteSelf      = New(ErrCodeAdminCannotDeleteSelf, "Cannot delete your own account")
	ErrAdminCannotDisableSelf     = New(ErrCodeAdminCannotDisableSelf, "Cannot disable your own account")
	ErrAdminCannotDemoteLastAdmin = New(ErrCodeAdminCannotDemoteLastAdmin, "Cannot demote the last platform administrator")
	ErrAdminCannotDeleteLastAdmin = New(ErrCodeAdminCannotDeleteLastAdmin, "Cannot delete the last platform administrator")
	ErrAdminCannotModifySystem    = New(ErrCodeAdminCannotModifySystem, "Cannot modify system user")
)
