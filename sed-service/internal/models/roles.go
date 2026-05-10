package models

const (
	RoleSedAdmin    = "sed_admin"
	RoleSedAuthor   = "sed_author"
	RoleSedApprover = "sed_approver"
	RoleSedViewer   = "sed_viewer"
)

// CanAdminSED справочники типов и маршрутов.
func CanAdminSED(c *Claims) bool {
	return c != nil && c.HasRole(RoleSedAdmin)
}

// CanAuthor документы.
func CanAuthor(c *Claims) bool {
	return c != nil && (c.HasRole(RoleSedAdmin) || c.HasRole(RoleSedAuthor))
}

// CanApprove согласование.
func CanApprove(c *Claims) bool {
	return c != nil && (c.HasRole(RoleSedAdmin) || c.HasRole(RoleSedApprover))
}

// CanViewSED чтение.
func CanViewSED(c *Claims) bool {
	return c != nil && (c.HasRole(RoleSedAdmin) || c.HasRole(RoleSedAuthor) || c.HasRole(RoleSedApprover) || c.HasRole(RoleSedViewer))
}
