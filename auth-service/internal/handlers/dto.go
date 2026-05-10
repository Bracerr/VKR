package handlers

// CreateTenantRequest тело POST /tenants.
type CreateTenantRequest struct {
	Code string `json:"code" binding:"required,min=2,max=64"`
	Name string `json:"name" binding:"required,max=255"`
}

// CreateUserRequest тело POST /users.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=1,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required"`
}

// UpdateRolesRequest тело PUT /users/:id/roles.
type UpdateRolesRequest struct {
	Roles []string `json:"roles" binding:"required,min=1"`
}

// TestLoginRequest тело POST /internal/test/login.
type TestLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
