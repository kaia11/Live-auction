package service

import (
	"strings"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

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
	store        *memoryStore
	tokenService *TokenService
	repo         repository.UserRepository
}

func NewUserService(tokenService *TokenService, repo repository.UserRepository) *UserService {
	return &UserService{store: sharedStore, tokenService: tokenService, repo: repo}
}

func (s *UserService) Login(phone string, password string) (LoginResult, error) {
	if strings.TrimSpace(phone) == "" || strings.TrimSpace(password) == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	userID := resolveLoginUserID(phone)
	user, err := s.loadUser(userID)
	if err != nil {
		return LoginResult{}, ErrUserNotFound
	}

	token, err := s.tokenService.Sign(user.ID, user.Role)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Token: token,
		User:  buildUserProfile(user, phone),
	}, nil
}

func (s *UserService) GetCurrentUser(authorization string) (UserProfile, error) {
	userID, err := s.GetCurrentUserID(authorization)
	if err != nil {
		return UserProfile{}, err
	}

	user, err := s.loadUser(userID)
	if err != nil {
		return UserProfile{}, ErrUserNotFound
	}

	return buildUserProfile(user, ""), nil
}

func (s *UserService) GetCurrentUserID(authorization string) (string, error) {
	claims, err := s.tokenService.Parse(authorization)
	if err != nil {
		return "", err
	}
	return claims.Sub, nil
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
	userID, err := s.GetCurrentUserID(authorization)
	if err != nil {
		return "", false
	}

	return userID, true
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

func (s *UserService) loadUser(userID string) (model.User, error) {
	if s.repo != nil {
		user, err := s.repo.GetByID(userID)
		if err == nil && user != nil && user.ID != "" {
			return *user, nil
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	user, ok := s.store.users[userID]
	if !ok {
		return model.User{}, ErrUserNotFound
	}
	return user, nil
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
