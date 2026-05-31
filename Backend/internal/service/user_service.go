package service

import (
	"strings"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
)

const mockTokenPrefix = "mock-token:"

type UserProfile struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type LoginResult struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

type UserService struct {
	store *memoryStore
}

func NewUserService() *UserService {
	return &UserService{store: sharedStore}
}

func (s *UserService) Login(phone string, password string) (LoginResult, error) {
	if strings.TrimSpace(phone) == "" || strings.TrimSpace(password) == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	userID := resolveLoginUserID(phone)
	user, ok := s.store.users[userID]
	if !ok {
		return LoginResult{}, ErrUserNotFound
	}

	return LoginResult{
		Token: buildMockToken(user.ID),
		User:  buildUserProfile(user, phone),
	}, nil
}

func (s *UserService) GetCurrentUser(authorization string) (UserProfile, error) {
	userID, err := s.GetCurrentUserID(authorization)
	if err != nil {
		return UserProfile{}, err
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	user, ok := s.store.users[userID]
	if !ok {
		return UserProfile{}, ErrUserNotFound
	}

	return buildUserProfile(user, ""), nil
}

func (s *UserService) GetCurrentUserID(authorization string) (string, error) {
	return parseMockToken(authorization)
}

func (s *UserService) RequireAnyRole(authorization string, allowedRoles ...string) (UserProfile, error) {
	user, err := s.GetCurrentUser(authorization)
	if err != nil {
		return UserProfile{}, err
	}

	for _, role := range allowedRoles {
		if user.Role == role {
			return user, nil
		}
	}

	return UserProfile{}, ErrForbiddenRole
}

func (s *UserService) TryGetCurrentUserID(authorization string) (string, bool) {
	userID, err := parseMockToken(authorization)
	if err != nil {
		return "", false
	}

	return userID, true
}

func buildMockToken(userID string) string {
	return mockTokenPrefix + userID
}

func parseMockToken(authorization string) (string, error) {
	token := strings.TrimSpace(authorization)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	if !strings.HasPrefix(token, mockTokenPrefix) {
		return "", ErrUnauthorizedToken
	}

	userID := strings.TrimPrefix(token, mockTokenPrefix)
	if userID == "" {
		return "", ErrUnauthorizedToken
	}

	return userID, nil
}

func buildUserProfile(user model.User, phone string) UserProfile {
	return UserProfile{
		ID:       user.ID,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Phone:    phone,
		Role:     user.Role,
	}
}

func resolveLoginUserID(phone string) string {
	normalized := strings.ToLower(strings.TrimSpace(phone))

	switch normalized {
	case "viewer", "user", "viewer_demo", "13800000001":
		return "user-001"
	case "anchor", "anchor_demo", "anchor_admin", "13900000001":
		return "anchor-001"
	case "admin", "admin_demo", "admin_root", "18800000001":
		return "admin-001"
	default:
		if strings.Contains(normalized, "anchor") {
			return "anchor-001"
		}
		if strings.Contains(normalized, "admin") {
			return "admin-001"
		}
		return "user-001"
	}
}

func IsAdminRole(role string) bool {
	return role == domain.UserRoleAnchor || role == domain.UserRoleAdmin
}
