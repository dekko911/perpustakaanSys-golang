package types

import "context"

type RoleUserStore interface {
	// relation method many to many with roles.
	GetUserWithRoleByUserID(ctx context.Context, userID string) (*User, error)
	GetUserAndRoleNames(ctx context.Context, userID string) (*User, map[string][]string, error)

	AssignRoleIntoUser(ctx context.Context, userID, roleID string) error
	DeleteRoleFromUser(ctx context.Context, userID, roleID string) error
}

type SetPayloadJSONRoleAndUserID struct {
	UserID string `json:"user_id" validate:"required"`
	RoleID string `json:"role_id" validate:"required"`
}
