package keycloak

import "testing"

func TestTenantPowerRoles_excludesSuperAdmin(t *testing.T) {
	roles := TenantPowerRoles()
	for _, r := range roles {
		if r == RoleSuperAdmin {
			t.Fatal("super_admin must not be in tenant power roles")
		}
	}
	if len(roles) != len(RealmRoles)-1 {
		t.Fatalf("expected %d roles, got %d", len(RealmRoles)-1, len(roles))
	}
}
