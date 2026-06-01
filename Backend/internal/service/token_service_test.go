package service

import (
	"testing"
	"time"
)

func TestTokenServiceSignAndParse(t *testing.T) {
	service := NewTokenService("secret-for-test")

	token, err := service.Sign("user-001", "viewer")
	if err != nil {
		t.Fatalf("expected sign success, got error: %v", err)
	}

	claims, err := service.Parse("Bearer " + token)
	if err != nil {
		t.Fatalf("expected parse success, got error: %v", err)
	}
	if claims.Sub != "user-001" {
		t.Fatalf("expected sub user-001, got %s", claims.Sub)
	}
	if claims.Role != "viewer" {
		t.Fatalf("expected role viewer, got %s", claims.Role)
	}
}

func TestTokenServiceRejectsExpiredToken(t *testing.T) {
	service := NewTokenService("secret-for-test")
	service.ttl = -1 * time.Minute

	token, err := service.Sign("user-001", "viewer")
	if err != nil {
		t.Fatalf("expected sign success, got error: %v", err)
	}

	if _, err := service.Parse(token); err == nil {
		t.Fatalf("expected expired token to be rejected")
	}
}

func TestTokenServiceRejectsWrongSignature(t *testing.T) {
	signer := NewTokenService("secret-a")
	parser := NewTokenService("secret-b")

	token, err := signer.Sign("user-001", "viewer")
	if err != nil {
		t.Fatalf("expected sign success, got error: %v", err)
	}

	if _, err := parser.Parse(token); err == nil {
		t.Fatalf("expected wrong signature token to be rejected")
	}
}
