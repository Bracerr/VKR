package keycloak

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
)

// BootstrapSuperAdmin создаёт пользователя с ролью super_admin в целевом realm (если включено в конфиге).
func BootstrapSuperAdmin(ctx context.Context, c *Client, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("empty username or password")
	}
	jwt, err := c.LoginAdmin(ctx)
	if err != nil {
		return err
	}
	token := jwt.AccessToken
	n, err := c.CountUsersByUsername(ctx, token, username)
	if err != nil {
		return err
	}
	enabled := true
	email := username + "@bootstrap.local"
	attrs := map[string][]string{"tenant_id": {"superadmin"}}
	var uid string
	if n > 0 {
		// Пользователь уже есть — снимаем «Account is not fully set up» (requiredActions / профиль).
		users, err := c.UsersByExactUsername(ctx, token, username)
		if err != nil || len(users) == 0 || users[0].ID == nil {
			return fmt.Errorf("bootstrap superadmin: user exists but cannot load")
		}
		uid = *users[0].ID
	} else {
		// Keycloak 24+ (declarative user profile): без first/last name password grant даёт
		// invalid_grant "Account is not fully set up".
		u := gocloak.User{
			Username:        gocloak.StringP(username),
			FirstName:       gocloak.StringP("Super"),
			LastName:        gocloak.StringP("Admin"),
			Enabled:         &enabled,
			Email:           gocloak.StringP(email),
			EmailVerified:   gocloak.BoolP(true),
			Attributes:      &attrs,
			RequiredActions: &[]string{},
		}
		uid, err = c.CreateUser(ctx, token, u)
		if err != nil {
			return fmt.Errorf("create superadmin: %w", err)
		}
		if err := c.SetUserPassword(ctx, token, uid, password, false); err != nil {
			_ = c.DeleteUser(ctx, token, uid)
			return err
		}
		assign := make([]gocloak.Role, 0, len(RealmRoles))
		for _, rn := range RealmRoles {
			role, err := c.RealmRole(ctx, token, rn)
			if err != nil {
				_ = c.DeleteUser(ctx, token, uid)
				return fmt.Errorf("bootstrap superadmin: load role %s: %w", rn, err)
			}
			assign = append(assign, *role)
		}
		if err := c.AddRealmRoleToUser(ctx, token, uid, assign); err != nil {
			_ = c.DeleteUser(ctx, token, uid)
			return fmt.Errorf("bootstrap superadmin: assign initial roles: %w", err)
		}
	}
	full, err := c.GetUserByID(ctx, token, uid)
	if err != nil {
		return fmt.Errorf("get superadmin after create: %w", err)
	}
	full.Email = gocloak.StringP(email)
	full.EmailVerified = gocloak.BoolP(true)
	full.Enabled = gocloak.BoolP(true)
	full.FirstName = gocloak.StringP("Super")
	full.LastName = gocloak.StringP("Admin")
	if full.Attributes == nil {
		full.Attributes = &map[string][]string{}
	}
	if vals, ok := (*full.Attributes)["tenant_id"]; !ok || len(vals) == 0 || vals[0] == "" {
		(*full.Attributes)["tenant_id"] = []string{"superadmin"}
	}
	full.RequiredActions = &[]string{}
	if err := c.UpdateUser(ctx, token, *full); err != nil {
		return fmt.Errorf("update superadmin (clear required actions): %w", err)
	}
	existing, err := c.GetRealmRolesOfUser(ctx, token, uid)
	if err != nil {
		return fmt.Errorf("bootstrap superadmin: get existing roles: %w", err)
	}
	have := make(map[string]struct{}, len(existing))
	for _, r := range existing {
		if r.Name != nil {
			have[*r.Name] = struct{}{}
		}
	}
	missing := make([]gocloak.Role, 0, len(RealmRoles))
	for _, rn := range RealmRoles {
		if _, ok := have[rn]; ok {
			continue
		}
		role, err := c.RealmRole(ctx, token, rn)
		if err != nil {
			return fmt.Errorf("bootstrap superadmin: realm role %s: %w", rn, err)
		}
		missing = append(missing, *role)
	}
	if len(missing) > 0 {
		if err := c.AddRealmRoleToUser(ctx, token, uid, missing); err != nil {
			return fmt.Errorf("bootstrap superadmin: assign missing roles: %w", err)
		}
	}
	return nil
}
