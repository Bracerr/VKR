package models

const (
	RoleProcAdmin    = "proc_admin"
	RoleProcBuyer    = "proc_buyer"
	RoleProcApprover = "proc_approver"
	RoleProcViewer   = "proc_viewer"
)

func CanAdminProc(c *Claims) bool {
	return c != nil && c.HasRole(RoleProcAdmin)
}

func CanBuy(c *Claims) bool {
	return c != nil && (c.HasRole(RoleProcAdmin) || c.HasRole(RoleProcBuyer))
}

func CanViewProc(c *Claims) bool {
	if c == nil {
		return false
	}
	return c.HasRole(RoleProcAdmin) || c.HasRole(RoleProcBuyer) || c.HasRole(RoleProcApprover) || c.HasRole(RoleProcViewer)
}

