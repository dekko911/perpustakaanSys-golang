package types

import "fmt"

// mock user store for test purpose
type MockUserStore struct{}

func (m MockUserStore) GetUsers() ([]*User, error) {
	return nil, nil
}

func (m MockUserStore) GetUserWithRolesByID(id string) (*User, error) {
	return nil, nil
}

func (m MockUserStore) GetUserWithRolesByEmail(email string) (*User, error) {
	return nil, fmt.Errorf("user not found")
}

func (m MockUserStore) CreateUser(*User) error {
	return nil
}

func (m MockUserStore) UpdateUser(id string, u *User) error {
	return nil
}

func (m MockUserStore) DeleteUser(id string) error {
	return nil
}

func (m MockUserStore) IncrementTokenVersion(id string) error {
	return nil
}

// mock role & user store for test purpose
type MockRoleUserStore struct{}

func (m MockRoleUserStore) GetUserWithRoleByUserID(userID string) (*User, error) {
	return nil, nil
}

func (m MockRoleUserStore) AssignRoleIntoUser(userID, roleID string) error {
	return nil
}

func (m MockRoleUserStore) DeleteRoleFromUser(userID, roleID string) error {
	return nil
}

// mock role store for test purpose
type MockRoleStore struct{}

func (m MockRoleStore) GetRoles() ([]*Role, error) {
	return nil, nil
}

func (m MockRoleStore) GetRoleByID(id string) (*Role, error) {
	return nil, nil
}

func (m MockRoleStore) GetRoleByName(name string) (*Role, error) {
	return nil, fmt.Errorf("role not found")
}

func (m MockRoleStore) CreateRole(*Role) error {
	return nil
}

func (m MockRoleStore) UpdateRole(id string, r Role) error {
	return nil
}

func (m MockRoleStore) DeleteRole(id string) error {
	return nil
}
