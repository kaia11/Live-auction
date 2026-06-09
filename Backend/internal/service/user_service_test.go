package service

import "testing"

func TestUserServiceRequireAnyRoleAllowsAnchor(t *testing.T) {
	tokenService := NewTokenService("secret-for-test")
	userService := NewUserService(tokenService, nil, nil, nil)

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
	userService := NewUserService(tokenService, nil, nil, nil)

	token, err := tokenService.Sign("user-001", "viewer")
	if err != nil {
		t.Fatalf("expected sign success, got error: %v", err)
	}

	if _, err := userService.RequireAnyRole("Bearer "+token, "anchor", "admin"); err == nil {
		t.Fatalf("expected viewer to be rejected for admin endpoints")
	}
}

func TestUserServiceRegisterAndLoginViewer(t *testing.T) {
	tokenService := NewTokenService("secret-for-test")
	userService := NewUserService(tokenService, nil, nil, nil)

	result, err := userService.Register("new_viewer", "pass123", "viewer")
	if err != nil {
		t.Fatalf("expected register success, got error: %v", err)
	}
	if result.User.Role != "viewer" {
		t.Fatalf("expected viewer role, got %s", result.User.Role)
	}

	loginResult, err := userService.Login("new_viewer", "pass123", "viewer")
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if loginResult.User.Username != "new_viewer" {
		t.Fatalf("expected username new_viewer, got %s", loginResult.User.Username)
	}
}

func TestUserServiceLoginRejectsCrossClientAccess(t *testing.T) {
	tokenService := NewTokenService("secret-for-test")
	userService := NewUserService(tokenService, nil, nil, nil)

	if _, err := userService.Login("silence", "111111", "anchor"); err == nil {
		t.Fatalf("expected viewer login to be rejected for anchor client")
	}
}
