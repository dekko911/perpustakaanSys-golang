package types

type RoleUserStore interface {
	// relation method many to many with roles.
	GetUserWithRoleByUserID(userID string) (*User, error)
	AssignRoleIntoUser(userID, roleID string) error
	DeleteRoleFromUser(userID, roleID string) error
}

type PayloadRoleUserID struct {
	UserID string `form:"user_id" validate:"required"`
	RoleID string `form:"role_id" validate:"required"`
}
