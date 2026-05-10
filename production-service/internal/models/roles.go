package models

const (
	RoleProdAdmin         = "prod_admin"
	RoleProdTechnologist  = "prod_technologist"
	RoleProdPlanner       = "prod_planner"
	RoleProdMaster        = "prod_master"
	RoleProdWorker        = "prod_worker"
	RoleProdQC            = "prod_qc"
	RoleProdViewer        = "prod_viewer"
)

// CanAdminPROD полный доступ.
func CanAdminPROD(c *Claims) bool {
	return c != nil && c.HasRole(RoleProdAdmin)
}

// CanTechnologist справочники производства.
func CanTechnologist(c *Claims) bool {
	return c != nil && (c.HasRole(RoleProdAdmin) || c.HasRole(RoleProdTechnologist))
}

// CanPlan заказы и смены.
func CanPlan(c *Claims) bool {
	return c != nil && (c.HasRole(RoleProdAdmin) || c.HasRole(RoleProdPlanner))
}

// CanMaster факт и закрытие операций.
func CanMaster(c *Claims) bool {
	return c != nil && (c.HasRole(RoleProdAdmin) || c.HasRole(RoleProdMaster))
}

// CanWorker наряды.
func CanWorker(c *Claims) bool {
	return c != nil && (c.HasRole(RoleProdAdmin) || c.HasRole(RoleProdMaster) || c.HasRole(RoleProdWorker))
}

// CanViewPROD чтение.
func CanViewPROD(c *Claims) bool {
	if c == nil {
		return false
	}
	return c.HasRole(RoleProdAdmin) || c.HasRole(RoleProdTechnologist) || c.HasRole(RoleProdPlanner) ||
		c.HasRole(RoleProdMaster) || c.HasRole(RoleProdWorker) || c.HasRole(RoleProdQC) || c.HasRole(RoleProdViewer)
}
