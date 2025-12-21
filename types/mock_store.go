package types

import (
	"context"
	"fmt"
)

// mock user store for test purpose
type MockUserStore struct{}

func (m MockUserStore) GetUsers(ctx context.Context) ([]*User, error) {
	return nil, nil
}

func (m MockUserStore) GetUserWithRolesByID(ctx context.Context, id string) (*User, error) {
	return nil, nil
}

func (m MockUserStore) GetUserWithRolesByEmail(ctx context.Context, email string) (*User, error) {
	return nil, fmt.Errorf("user not found")
}

func (m MockUserStore) CreateUser(ctx context.Context, u *User) error {
	return nil
}

func (m MockUserStore) UpdateUser(ctx context.Context, id string, u *User) error {
	return nil
}

func (m MockUserStore) DeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m MockUserStore) IncrementTokenVersion(ctx context.Context, id, token string) error {
	return nil
}

// mock role & user store for test purpose
type MockRoleUserStore struct{}

func (m MockRoleUserStore) GetUserWithRoleByUserID(ctx context.Context, userID string) (*User, error) {
	return nil, nil
}

func (m MockRoleUserStore) AssignRoleIntoUser(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m MockRoleUserStore) DeleteRoleFromUser(ctx context.Context, userID, roleID string) error {
	return nil
}

// mock role store for test purpose
type MockRoleStore struct{}

func (m MockRoleStore) GetRoles(ctx context.Context) ([]*Role, error) {
	return nil, nil
}

func (m MockRoleStore) GetRoleByID(ctx context.Context, id string) (*Role, error) {
	return nil, nil
}

func (m MockRoleStore) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	return nil, fmt.Errorf("role not found")
}

func (m MockRoleStore) CreateRole(ctx context.Context, r Role) error {
	return nil
}

func (m MockRoleStore) UpdateRole(ctx context.Context, id string, r Role) error {
	return nil
}

func (m MockRoleStore) DeleteRole(ctx context.Context, id string) error {
	return nil
}

type MockMemberStore struct{}

func (m MockMemberStore) GetMembers(ctx context.Context) ([]*Member, error) {
	return nil, nil
}

func (m MockMemberStore) GetMemberByID(ctx context.Context, id string) (*Member, error) {
	return nil, nil
}

func (m MockMemberStore) GetMemberByNama(ctx context.Context, nama string) (*Member, error) {
	return nil, fmt.Errorf("member not found")
}

func (m MockMemberStore) GetMemberByNoTelepon(ctx context.Context, no_phone string) (*Member, error) {
	return nil, fmt.Errorf("member not found")
}

func (mm MockMemberStore) CreateMember(ctx context.Context, m *Member) error {
	return nil
}

func (mm MockMemberStore) UpdateMember(ctx context.Context, id string, m *Member) error {
	return nil
}

func (m MockMemberStore) DeleteMember(ctx context.Context, id string) error {
	return nil
}

type MockCirculationStore struct{}

func (m MockCirculationStore) GetCirculations(ctx context.Context) ([]*Circulation, error) {
	return nil, nil
}

func (m MockCirculationStore) GetCirculationByID(ctx context.Context, id string) (*Circulation, error) {
	return nil, nil
}

func (m MockCirculationStore) GetCirculationByPeminjam(ctx context.Context, borrowerName string) (*Circulation, error) {
	return nil, fmt.Errorf("circulation not found")
}

func (m MockCirculationStore) CreateCirculation(ctx context.Context, c *Circulation) error {
	return nil
}

func (m MockCirculationStore) UpdateCirculation(ctx context.Context, id string, c *Circulation) error {
	return nil
}

func (m MockCirculationStore) DeleteCirculation(ctx context.Context, id string) error {
	return nil
}

type MockBookStore struct{}

func (m MockBookStore) GetBooks(ctx context.Context) ([]*Book, error) {
	return nil, nil
}

func (m MockBookStore) GetBookByID(ctx context.Context, id string) (*Book, error) {
	return nil, nil
}

func (m MockBookStore) GetBookByJudulBuku(ctx context.Context, judulBuku string) (*Book, error) {
	return nil, fmt.Errorf("book not found")
}

func (m MockBookStore) CreateBook(ctx context.Context, b *Book) error {
	return nil
}

func (m MockBookStore) UpdateBook(ctx context.Context, id string, b *Book) error {
	return nil
}

func (m MockBookStore) DeleteBook(ctx context.Context, id string) error {
	return nil
}
