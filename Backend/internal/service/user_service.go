package service

import (
	"strings"

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

	user, ok := s.store.users["user-001"]
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
