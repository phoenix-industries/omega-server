package auth

import "errors"

type Role string

const (
	RoleRoot    Role = "root"
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleMember  Role = "member"
)

func (r Role) String() string {
	return string(r)
}

func (r *Role) Scan(value any) error {
	switch value.(string) {
	case "root":
		*r = RoleRoot
	case "admin":
		*r = RoleAdmin
	case "manager":
		*r = RoleManager
	case "member":
		*r = RoleMember
	default:
		return errors.New("invalid user role")
	}
	return nil
}

func AllowedRole(current Role, required Role) bool {
	switch current {
	case required, RoleRoot:
		return true
	case RoleAdmin:
		return required == RoleManager || required == RoleMember
	case RoleManager:
		return required == RoleMember
	default:
		return false
	}
}

func (r Role) Allowed(required Role) bool {
	return AllowedRole(r, required)
}
