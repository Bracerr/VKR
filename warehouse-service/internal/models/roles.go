package models

const (
	RoleWarehouseAdmin = "warehouse_admin"
	RoleStorekeeper    = "storekeeper"
	RoleWarehouseView  = "warehouse_viewer"
)

// CanAdminCatalog CRUD справочников.
func CanAdminCatalog(c *Claims) bool {
	return c != nil && c.HasRole(RoleWarehouseAdmin)
}

// CanOperate складские операции и резервы.
func CanOperate(c *Claims) bool {
	return c != nil && (c.HasRole(RoleWarehouseAdmin) || c.HasRole(RoleStorekeeper))
}

// CanView чтение.
func CanView(c *Claims) bool {
	return c != nil && (c.HasRole(RoleWarehouseAdmin) || c.HasRole(RoleStorekeeper) || c.HasRole(RoleWarehouseView))
}
