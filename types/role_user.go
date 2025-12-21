package types

import "context"

type RoleUserStore interface {
	// relation method many to many with roles.
	GetUserWithRoleByUserID(ctx context.Context, userID string) (*User, error)
	AssignRoleIntoUser(ctx context.Context, userID, roleID string) error
	DeleteRoleFromUser(ctx context.Context, userID, roleID string) error
}

type SetPayloadRoleAndUserID struct {
	UserID string `form:"user_id" validate:"required"`
	RoleID string `form:"role_id" validate:"required"`
}
