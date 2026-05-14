package models

const (
	RoleSalesAdmin    = "sales_admin"
	RoleSalesManager  = "sales_manager"
	RoleSalesApprover = "sales_approver"
	RoleSalesViewer   = "sales_viewer"
)

func CanAdminSales(c *Claims) bool {
	return c != nil && c.HasRole(RoleSalesAdmin)
}

func CanManageSales(c *Claims) bool {
	return c != nil && (c.HasRole(RoleSalesAdmin) || c.HasRole(RoleSalesManager))
}

func CanViewSales(c *Claims) bool {
	if c == nil {
		return false
	}
	return c.HasRole(RoleSalesAdmin) || c.HasRole(RoleSalesManager) || c.HasRole(RoleSalesApprover) || c.HasRole(RoleSalesViewer)
}

