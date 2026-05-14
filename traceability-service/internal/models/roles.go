package models

// Traceability — read-only сервис, но ограничим доступ "viewer+" по уже существующим ролям доменов.
// Это позволяет не добавлять новый realm role в Keycloak для MVP.

var viewerOrHigherRoles = []string{
	// enterprise
	"ent_admin",
	// warehouse
	"warehouse_admin", "storekeeper", "warehouse_viewer",
	// sales
	"sales_admin", "sales_manager", "sales_viewer",
	// procurement
	"proc_admin", "proc_buyer", "proc_viewer",
	// production
	"prod_admin", "prod_technologist", "prod_planner", "prod_master", "prod_worker", "prod_qc", "prod_viewer",
	// sed
	"sed_admin", "sed_author", "sed_approver", "sed_viewer",
}

func CanViewTrace(c *Claims) bool {
	if c == nil {
		return false
	}
	for _, r := range viewerOrHigherRoles {
		if c.HasRole(r) {
			return true
		}
	}
	return false
}

