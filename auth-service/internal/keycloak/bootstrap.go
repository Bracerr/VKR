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
		role, err := c.RealmRole(ctx, token, RoleSuperAdmin)
		if err != nil {
			_ = c.DeleteUser(ctx, token, uid)
			return err
		}
		if err := c.AddRealmRoleToUser(ctx, token, uid, []gocloak.Role{*role}); err != nil {
			_ = c.DeleteUser(ctx, token, uid)
			return err
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
	full.RequiredActions = &[]string{}
	if err := c.UpdateUser(ctx, token, *full); err != nil {
		return fmt.Errorf("update superadmin (clear required actions): %w", err)
	}
	hasSuper := false
	if roles, err := c.GetRealmRolesOfUser(ctx, token, uid); err == nil {
		for _, r := range roles {
			if r.Name != nil && *r.Name == RoleSuperAdmin {
				hasSuper = true
				break
			}
		}
	}
	if !hasSuper {
		role, err := c.RealmRole(ctx, token, RoleSuperAdmin)
		if err != nil {
			return fmt.Errorf("bootstrap superadmin: realm role: %w", err)
		}
		if err := c.AddRealmRoleToUser(ctx, token, uid, []gocloak.Role{*role}); err != nil {
			return fmt.Errorf("bootstrap superadmin: assign role: %w", err)
		}
	}
	return nil
}
