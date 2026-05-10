package models

// Claims — извлечённые из JWT данные для контекста запроса.
type Claims struct {
	Sub       string   `json:"sub"`
	Username  string   `json:"preferred_username"`
	Email     string   `json:"email"`
	TenantID  string   `json:"tenant_id"`
	RealmRoles []string `json:"realm_roles"`
}

// HasRole проверяет наличие роли realm.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.RealmRoles {
		if r == role {
			return true
		}
	}
	return false
}
