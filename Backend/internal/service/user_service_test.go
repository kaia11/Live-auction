package service

import "testing"

func TestUserServiceRequireAnyRoleAllowsAnchor(t *testing.T) {
	tokenService := NewTokenService("secret-for-test")
	userService := NewUserService(tokenService, nil)

	token, err := tokenService.Sign("anchor-001", "anchor")
	if err != nil {
		t.Fatalf("expected sign success, got error: %v", err)
	}

	user, err := userService.RequireAnyRole("Bearer "+token, "anchor", "admin")
	if err != nil {
		t.Fatalf("expected anchor to pass role check, got error: %v", err)
	}
	if user.ID != "anchor-001" {
		t.Fatalf("expected anchor-001, got %s", user.ID)
	}
}

func TestUserServiceRequireAnyRoleRejectsViewer(t *testing.T) {
	tokenService := NewTokenService("secret-for-test")
	userService := NewUserService(tokenService, nil)

	token, err := tokenService.Sign("user-001", "viewer")
	if err != nil {
		t.Fatalf("expected sign success, got error: %v", err)
	}

	if _, err := userService.RequireAnyRole("Bearer "+token, "anchor", "admin"); err == nil {
		t.Fatalf("expected viewer to be rejected for admin endpoints")
	}
}
