package types

import "fmt"

// mock user store for test purpose
type MockUserStore struct{}

func (m MockUserStore) GetUsers() ([]*User, error) {
	return nil, nil
}

func (m MockUserStore) GetUsersBySearch(search string) []*User {
	return nil
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

func (m MockRoleStore) CreateRole(Role) error {
	return nil
}

func (m MockRoleStore) UpdateRole(id string, r Role) error {
	return nil
}

func (m MockRoleStore) DeleteRole(id string) error {
	return nil
}

type MockMemberStore struct{}

func (m MockMemberStore) GetMembers() ([]*Member, error) {
	return nil, nil
}

func (m MockMemberStore) GetMemberByID(id string) (*Member, error) {
	return nil, nil
}

func (m MockMemberStore) GetMemberByNama(nama string) (*Member, error) {
	return nil, fmt.Errorf("member not found")
}

func (m MockMemberStore) GetMemberByNoTelepon(no_phone string) (*Member, error) {
	return nil, fmt.Errorf("member not found")
}

func (m MockMemberStore) CreateMember(*Member) error {
	return nil
}

func (m MockMemberStore) UpdateMember(id string, mem *Member) error {
	return nil
}

func (m MockMemberStore) DeleteMember(id string) error {
	return nil
}

type MockCirculationStore struct{}

func (m MockCirculationStore) GetCirculations() ([]*Circulation, error) {
	return nil, nil
}

func (m MockCirculationStore) GetCirculationByID(id string) (*Circulation, error) {
	return nil, nil
}

func (m MockCirculationStore) GetCirculationByPeminjam(borrowerName string) (*Circulation, error) {
	return nil, fmt.Errorf("circulation not found")
}

func (m MockCirculationStore) CreateCirculation(*Circulation) error {
	return nil
}

func (m MockCirculationStore) UpdateCirculation(id string, c *Circulation) error {
	return nil
}

func (m MockCirculationStore) DeleteCirculation(id string) error {
	return nil
}

type MockBookStore struct{}

func (m MockBookStore) GetBooks() ([]*Book, error) {
	return nil, nil
}

func (m MockBookStore) GetBookByID(id string) (*Book, error) {
	return nil, nil
}

func (m MockBookStore) GetBookByJudulBuku(judulBuku string) (*Book, error) {
	return nil, fmt.Errorf("book not found")
}

func (m MockBookStore) CreateBook(*Book) error {
	return nil
}

func (m MockBookStore) UpdateBook(id string, b *Book) error {
	return nil
}

func (m MockBookStore) DeleteBook(id string) error {
	return nil
}
