package types

import "time"

type Role struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleStore interface {
	GetRoles() ([]*Role, error)
	GetRoleByID(id string) (*Role, error)
	GetRoleByName(name string) (*Role, error)
	CreateRole(*Role) error
	UpdateRole(id string, r Role) error
	DeleteRole(id string) error
}

type PayloadRole struct {
	Name string `form:"name" validate:"required,min=3"`
}

type PayloadUpdateRole struct {
	Name string `form:"name" validate:"omitempty,required,min=3"`
}

// relation many to many with users.
type Roles []Role
