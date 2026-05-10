package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/industrial-sed/auth-service/internal/keycloak"
	"github.com/industrial-sed/auth-service/internal/mocks"
	"github.com/industrial-sed/auth-service/internal/models"
	"github.com/industrial-sed/auth-service/internal/usecases"
)

func adminJWT() *gocloak.JWT {
	return &gocloak.JWT{AccessToken: "admin-token", ExpiresIn: 120}
}

func TestUserUC_Create_happyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	kc := mocks.NewMockKeycloakClient(ctrl)
	tr := mocks.NewMockTenantRepository(ctrl)
	ur := mocks.NewMockUserCacheRepository(ctrl)
	n := mocks.NewMockNotifier(ctrl)
	uc := usecases.NewUserUC(kc, tr, ur, n)

	ctx := context.Background()
	actor := &models.Claims{TenantID: "rom", RealmRoles: []string{keycloak.RoleEntAdmin}}

	kc.EXPECT().LoginAdmin(ctx).Return(adminJWT(), nil).AnyTimes()
	tr.EXPECT().GetByCode(ctx, "rom").Return(&models.Tenant{Code: "rom", KeycloakGroupID: "g1"}, nil)
	kc.EXPECT().CountUsersByUsername(ctx, "admin-token", "ivan@rom").Return(0, nil)
	kc.EXPECT().CreateUser(ctx, "admin-token", gomock.AssignableToTypeOf(gocloak.User{})).Return("uid-1", nil)
	kc.EXPECT().SetUserPassword(ctx, "admin-token", "uid-1", gomock.Any(), false).Return(nil)
	kc.EXPECT().AddUserToGroup(ctx, "admin-token", "uid-1", "g1").Return(nil)
	kc.EXPECT().RealmRole(ctx, "admin-token", keycloak.RoleEngineer).Return(&gocloak.Role{Name: gocloak.StringP(keycloak.RoleEngineer)}, nil)
	kc.EXPECT().AddRealmRoleToUser(ctx, "admin-token", "uid-1", gomock.Any()).Return(nil)
	ur.EXPECT().Upsert(ctx, gomock.AssignableToTypeOf(&models.UserCache{})).Return(nil)
	n.EXPECT().NotifyUserCreated(ctx, gomock.Any()).Return(nil)

	id, pw, err := uc.CreateUser(ctx, actor, "ivan", "i@rom.ru", keycloak.RoleEngineer)
	require.NoError(t, err)
	require.Equal(t, "uid-1", id)
	require.NotEmpty(t, pw)
}

func TestUserUC_Create_conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	kc := mocks.NewMockKeycloakClient(ctrl)
	tr := mocks.NewMockTenantRepository(ctrl)
	ur := mocks.NewMockUserCacheRepository(ctrl)
	n := mocks.NewMockNotifier(ctrl)
	uc := usecases.NewUserUC(kc, tr, ur, n)
	ctx := context.Background()
	actor := &models.Claims{TenantID: "rom", RealmRoles: []string{keycloak.RoleEntAdmin}}

	kc.EXPECT().LoginAdmin(ctx).Return(adminJWT(), nil).AnyTimes()
	tr.EXPECT().GetByCode(ctx, "rom").Return(&models.Tenant{KeycloakGroupID: "g1"}, nil)
	kc.EXPECT().CountUsersByUsername(ctx, "admin-token", "ivan@rom").Return(1, nil)

	_, _, err := uc.CreateUser(ctx, actor, "ivan", "i@rom.ru", keycloak.RoleEngineer)
	require.ErrorIs(t, err, usecases.ErrConflict)
}

func TestUserUC_Create_keycloakError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	kc := mocks.NewMockKeycloakClient(ctrl)
	tr := mocks.NewMockTenantRepository(ctrl)
	ur := mocks.NewMockUserCacheRepository(ctrl)
	n := mocks.NewMockNotifier(ctrl)
	uc := usecases.NewUserUC(kc, tr, ur, n)
	ctx := context.Background()
	actor := &models.Claims{TenantID: "rom", RealmRoles: []string{keycloak.RoleEntAdmin}}

	kc.EXPECT().LoginAdmin(ctx).Return(adminJWT(), nil).AnyTimes()
	tr.EXPECT().GetByCode(ctx, "rom").Return(&models.Tenant{KeycloakGroupID: "g1"}, nil)
	kc.EXPECT().CountUsersByUsername(ctx, "admin-token", "ivan@rom").Return(0, nil)
	kc.EXPECT().CreateUser(ctx, "admin-token", gomock.Any()).Return("", errors.New("kc down"))

	_, _, err := uc.CreateUser(ctx, actor, "ivan", "i@rom.ru", keycloak.RoleEngineer)
	require.Error(t, err)
}

func TestUserUC_Create_notifierError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	kc := mocks.NewMockKeycloakClient(ctrl)
	tr := mocks.NewMockTenantRepository(ctrl)
	ur := mocks.NewMockUserCacheRepository(ctrl)
	n := mocks.NewMockNotifier(ctrl)
	uc := usecases.NewUserUC(kc, tr, ur, n)
	ctx := context.Background()
	actor := &models.Claims{TenantID: "rom", RealmRoles: []string{keycloak.RoleEntAdmin}}

	kc.EXPECT().LoginAdmin(ctx).Return(adminJWT(), nil).AnyTimes()
	tr.EXPECT().GetByCode(ctx, "rom").Return(&models.Tenant{KeycloakGroupID: "g1"}, nil)
	kc.EXPECT().CountUsersByUsername(ctx, "admin-token", "ivan@rom").Return(0, nil)
	kc.EXPECT().CreateUser(ctx, "admin-token", gomock.Any()).Return("uid-1", nil)
	kc.EXPECT().SetUserPassword(ctx, "admin-token", "uid-1", gomock.Any(), false).Return(nil)
	kc.EXPECT().AddUserToGroup(ctx, "admin-token", "uid-1", "g1").Return(nil)
	kc.EXPECT().RealmRole(ctx, "admin-token", keycloak.RoleEngineer).Return(&gocloak.Role{Name: gocloak.StringP(keycloak.RoleEngineer)}, nil)
	kc.EXPECT().AddRealmRoleToUser(ctx, "admin-token", "uid-1", gomock.Any()).Return(nil)
	ur.EXPECT().Upsert(ctx, gomock.Any()).Return(nil)
	n.EXPECT().NotifyUserCreated(ctx, gomock.Any()).Return(errors.New("kafka unavailable"))
	kc.EXPECT().DeleteUser(ctx, "admin-token", "uid-1").Return(nil)
	ur.EXPECT().Delete(ctx, "uid-1").Return(nil)

	_, _, err := uc.CreateUser(ctx, actor, "ivan", "i@rom.ru", keycloak.RoleEngineer)
	require.Error(t, err)
}
