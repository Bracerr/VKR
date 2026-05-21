package keycloak

// Роли realm (RBAC на бэкенде).
const (
	RoleSuperAdmin = "super_admin"
	RoleEntAdmin   = "ent_admin"
	RoleApprover   = "approver"
	RoleEngineer   = "engineer"
	RoleViewer     = "viewer"

	// Склад (warehouse-service, JWT realm_access.roles).
	RoleWarehouseAdmin   = "warehouse_admin"
	RoleStorekeeper      = "storekeeper"
	RoleWarehouseViewer  = "warehouse_viewer"

	// СЭД (sed-service, JWT realm_access.roles).
	RoleSedAdmin    = "sed_admin"
	RoleSedAuthor   = "sed_author"
	RoleSedApprover = "sed_approver"
	RoleSedViewer   = "sed_viewer"

	// СЭД: ACL по типам документов.
	RoleDocReadProcurement  = "doc_read_procurement"
	RoleDocReadSales        = "doc_read_sales"
	RoleDocReadProduction   = "doc_read_production"
	RoleDocReadWarehouse    = "doc_read_warehouse"
	RoleDocReadFinance      = "doc_read_finance"
	RoleDocWriteProcurement = "doc_write_procurement"
	RoleDocWriteSales       = "doc_write_sales"
	RoleDocWriteProduction  = "doc_write_production"

	// Производство (production-service).
	RoleProdAdmin        = "prod_admin"
	RoleProdTechnologist = "prod_technologist"
	RoleProdPlanner      = "prod_planner"
	RoleProdMaster       = "prod_master"
	RoleProdWorker       = "prod_worker"
	RoleProdQC           = "prod_qc"
	RoleProdViewer       = "prod_viewer"

	// Закупки (procurement-service).
	RoleProcAdmin    = "proc_admin"
	RoleProcBuyer    = "proc_buyer"
	RoleProcApprover = "proc_approver"
	RoleProcViewer   = "proc_viewer"

	// Продажи (sales-service).
	RoleSalesAdmin    = "sales_admin"
	RoleSalesManager  = "sales_manager"
	RoleSalesApprover = "sales_approver"
	RoleSalesViewer   = "sales_viewer"
)

// RealmRoles список базовых ролей для bootstrap.
var RealmRoles = []string{
	RoleSuperAdmin,
	RoleEntAdmin,
	RoleApprover,
	RoleEngineer,
	RoleViewer,
	RoleWarehouseAdmin,
	RoleStorekeeper,
	RoleWarehouseViewer,
	RoleSedAdmin,
	RoleSedAuthor,
	RoleSedApprover,
	RoleSedViewer,
	RoleDocReadProcurement,
	RoleDocReadSales,
	RoleDocReadProduction,
	RoleDocReadWarehouse,
	RoleDocReadFinance,
	RoleDocWriteProcurement,
	RoleDocWriteSales,
	RoleDocWriteProduction,
	RoleProdAdmin,
	RoleProdTechnologist,
	RoleProdPlanner,
	RoleProdMaster,
	RoleProdWorker,
	RoleProdQC,
	RoleProdViewer,
	RoleProcAdmin,
	RoleProcBuyer,
	RoleProcApprover,
	RoleProcViewer,
	RoleSalesAdmin,
	RoleSalesManager,
	RoleSalesApprover,
	RoleSalesViewer,
}
