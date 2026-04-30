package authn

import "testing"

func TestClaimsFromPayloadPromotesAdminRole(t *testing.T) {
	cfg := Config{
		RoleClaim:   "groups",
		DefaultRole: RoleUser,
		AdminRoles:  roleSet("order-service-admin"),
		UserRoles:   roleSet("order-service-user"),
	}

	claims := claimsFromPayload(map[string]any{
		"sub":    "123",
		"email":  "admin@example.test",
		"groups": []any{"order-service-user", "order-service-admin"},
	}, cfg)

	if claims.Role != RoleAdmin {
		t.Fatalf("expected admin role, got %q", claims.Role)
	}
}

func TestClaimsFromPayloadDefaultsToUser(t *testing.T) {
	cfg := Config{
		RoleClaim:   "groups",
		DefaultRole: RoleUser,
		AdminRoles:  roleSet("order-service-admin"),
		UserRoles:   roleSet("order-service-user"),
	}

	claims := claimsFromPayload(map[string]any{
		"sub": "123",
	}, cfg)

	if claims.Role != RoleUser {
		t.Fatalf("expected user role, got %q", claims.Role)
	}
}
