package authn

import "testing"

func TestRoleLimiterUsesRoleLimits(t *testing.T) {
	limiter := newRoleLimiter(Config{
		UserLimitPerSecond:  1,
		AdminLimitPerSecond: 2,
	})

	user := Claims{Subject: "u1", Role: RoleUser}
	if !limiter.Allow(user) {
		t.Fatal("expected first user request to be allowed")
	}
	if limiter.Allow(user) {
		t.Fatal("expected second user request to be rate limited")
	}

	admin := Claims{Subject: "a1", Role: RoleAdmin}
	if !limiter.Allow(admin) || !limiter.Allow(admin) {
		t.Fatal("expected first two admin requests to be allowed")
	}
	if limiter.Allow(admin) {
		t.Fatal("expected third admin request to be rate limited")
	}
}
