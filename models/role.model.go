package models

type RoleName string

const (
	RoleSuperAdmin RoleName = "SUPER_ADMIN"
	RoleAdmin      RoleName = "ADMIN"
	RoleUser       RoleName = "USER"
	RoleCommAdmin  RoleName = "COMM_ADMIN"
	RoleChatAdmin  RoleName = "CHAT_ADMIN"
)

type Role struct {
	ID          int64    `json:"id" db:"id"`
	Name        RoleName `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
}

func NewRole(name RoleName, description string) Role {
	return Role{
		Name:        name,
		Description: description,
	}
}

func (r RoleName) String() string {
	return string(r)
}
