package models

import "time"

// Tenant — локальный кэш предприятия (источник истины группы — Keycloak).
type Tenant struct {
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	KeycloakGroupID  string    `json:"keycloak_group_id"`
	CreatedAt        time.Time `json:"created_at"`
}
