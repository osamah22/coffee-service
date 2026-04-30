package authn

import (
	"encoding/json"
	"strings"
)

type Role string

const (
	RoleGuest Role = "guest"
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Claims struct {
	Subject string
	Email   string
	Role    Role
	Roles   []string
}

func claimsFromPayload(payload map[string]any, cfg Config) Claims {
	roles := extractRoles(payload[cfg.RoleClaim])
	role := cfg.DefaultRole
	if hasAnyRole(roles, cfg.AdminRoles) {
		role = RoleAdmin
	} else if hasAnyRole(roles, cfg.UserRoles) {
		role = RoleUser
	}

	return Claims{
		Subject: stringClaim(payload, "sub"),
		Email:   stringClaim(payload, "email"),
		Role:    role,
		Roles:   roles,
	}
}

func stringClaim(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func extractRoles(raw any) []string {
	switch value := raw.(type) {
	case string:
		return splitRoleString(value)
	case []string:
		return normalizeRoles(value)
	case []any:
		roles := make([]string, 0, len(value))
		for _, item := range value {
			if role, ok := item.(string); ok {
				roles = append(roles, role)
			}
		}
		return normalizeRoles(roles)
	case json.RawMessage:
		var values []string
		if json.Unmarshal(value, &values) == nil {
			return normalizeRoles(values)
		}
	}
	return nil
}

func splitRoleString(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	return normalizeRoles(fields)
}

func normalizeRoles(values []string) []string {
	roles := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		roles = append(roles, value)
	}
	return roles
}

func hasAnyRole(roles []string, allowed map[string]struct{}) bool {
	for _, role := range roles {
		if _, ok := allowed[role]; ok {
			return true
		}
	}
	return false
}
