package authn

import "testing"

func TestRoleLimiterUsesRoleLimits(t *testing.T) {
	limiter := newRoleLimiter(Config{
		GuestLimitPerSecond: 1,
		UserLimitPerSecond:  1,
		AdminLimitPerSecond: 2,
	})

	guest := Claims{Subject: "guest:127.0.0.1", Role: RoleGuest}
	if !limiter.Allow(guest) {
		t.Fatal("expected first guest request to be allowed")
	}
	if limiter.Allow(guest) {
		t.Fatal("expected second guest request to be rate limited")
	}

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

func TestGuestClaimsUseIPAsSubject(t *testing.T) {
	claims := guestClaims("192.0.2.10")
	if claims.Subject != "guest:192.0.2.10" {
		t.Fatalf("expected guest subject to include IP, got %q", claims.Subject)
	}
	if claims.Role != RoleGuest {
		t.Fatalf("expected guest role, got %q", claims.Role)
	}
}
