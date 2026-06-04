package service

import (
	"fmt"
	"strings"
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserProfile struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
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
	roomRepo     repository.RoomRepository
	cache        *userCache
}

func NewUserService(tokenService *TokenService, repo repository.UserRepository, roomRepo repository.RoomRepository, redisClient *realtime.Client) *UserService {
	return &UserService{
		store:        sharedStore,
		tokenService: tokenService,
		repo:         repo,
		roomRepo:     roomRepo,
		cache:        newUserCache(redisClient),
	}
}

func (s *UserService) Login(username string, password string, clientType string) (LoginResult, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	allowedRoles, err := allowedRolesForClient(clientType)
	if err != nil {
		return LoginResult{}, err
	}

	user, err := s.loadUserByUsername(username)
	if err != nil {
		return LoginResult{}, ErrUserNotFound
	}
	passwordMatched, needsRehash := passwordMatches(user.Password, password)
	if !passwordMatched || !hasAllowedRole(user.Role, allowedRoles...) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if needsRehash {
		if err := s.rehashPassword(&user, password); err != nil {
			return LoginResult{}, err
		}
	}

	token, err := s.tokenService.Sign(user.ID, user.Role)
	if err != nil {
		return LoginResult{}, err
	}
	_ = s.cacheUserProfile(user)

	return LoginResult{
		Token: token,
		User:  buildUserProfile(user),
	}, nil
}

func (s *UserService) Register(username string, password string, clientType string) (LoginResult, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	role, err := registrationRoleForClient(clientType)
	if err != nil {
		return LoginResult{}, err
	}

	if _, err := s.loadUserByUsername(username); err == nil {
		return LoginResult{}, ErrUsernameTaken
	}

	user := model.User{
		ID:       nextUserID(role),
		Username: username,
		Password: "",
		Nickname: username,
		Avatar:   defaultAvatarForRole(role),
		Role:     role,
	}
	if user.Password, err = hashPassword(password); err != nil {
		return LoginResult{}, err
	}
	if err := s.saveUser(user); err != nil {
		return LoginResult{}, err
	}
	if role == domain.UserRoleAnchor {
		if err := s.createAnchorRoom(user); err != nil {
			return LoginResult{}, err
		}
	}

	token, err := s.tokenService.Sign(user.ID, user.Role)
	if err != nil {
		return LoginResult{}, err
	}
	_ = s.cacheUserProfile(user)

	return LoginResult{
		Token: token,
		User:  buildUserProfile(user),
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

	return buildUserProfile(user), nil
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

func buildUserProfile(user model.User) UserProfile {
	return UserProfile{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Role:     user.Role,
	}
}

func (s *UserService) loadUser(userID string) (model.User, error) {
	if s.cache != nil {
		cachedUser, ok, err := s.cache.GetByID(userID)
		if err == nil && ok && cachedUser.ID != "" {
			return cachedUser, nil
		}
	}

	if s.repo != nil {
		user, err := s.repo.GetByID(userID)
		if err == nil && user != nil && user.ID != "" {
			_ = s.cacheUserProfile(*user)
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

func (s *UserService) loadUserByUsername(username string) (model.User, error) {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if normalized == "" {
		return model.User{}, ErrUserNotFound
	}

	if s.repo != nil {
		user, err := s.repo.GetByUsername(normalized)
		if err == nil && user != nil && user.ID != "" {
			_ = s.cacheUserProfile(*user)
			return *user, nil
		}
	}

	user := s.store.GetUserByUsername(normalized)
	if user.ID == "" {
		return model.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) saveUser(user model.User) error {
	if s.repo != nil {
		if err := s.repo.Create(user); err != nil {
			return err
		}
		return s.cacheUserProfile(user)
	}
	s.store.SaveUser(user)
	return nil
}

func (s *UserService) WarmUserCacheAndMigratePasswords() error {
	if s.repo == nil {
		return nil
	}

	users, err := s.repo.List()
	if err != nil {
		return err
	}

	for i := range users {
		user := users[i]
		if !isBcryptHash(user.Password) {
			if err := s.rehashPassword(&user, user.Password); err != nil {
				return err
			}
		}
		if err := s.cacheUserProfile(user); err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) rehashPassword(user *model.User, plaintext string) error {
	if s.repo == nil {
		user.Password = plaintext
		return nil
	}

	hash, err := hashPassword(plaintext)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePasswordHash(user.ID, hash); err != nil {
		return err
	}
	user.Password = hash
	return s.cacheUserProfile(*user)
}

func (s *UserService) cacheUserProfile(user model.User) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Set(user)
}

func (s *UserService) createAnchorRoom(user model.User) error {
	room := model.LiveRoom{
		ID:               fmt.Sprintf("room-%d", time.Now().UnixNano()),
		Title:            fmt.Sprintf("%s的直播间", user.Nickname),
		CoverImage:       "",
		VideoURL:         "",
		Status:           domain.RoomStatusOffline,
		AnchorUserID:     user.ID,
		AnchorName:       user.Nickname,
		OnlineCount:      0,
		Thumbnail:        "",
		CurrentSessionID: "",
	}

	s.store.mu.Lock()
	s.store.rooms[room.ID] = room
	s.store.roomItems[room.ID] = []string{}
	s.store.mu.Unlock()

	if s.roomRepo != nil {
		return s.roomRepo.SaveRoom(room)
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func passwordMatches(storedPassword string, password string) (bool, bool) {
	if isBcryptHash(storedPassword) {
		err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		return err == nil, false
	}
	return storedPassword == password, true
}

func isBcryptHash(value string) bool {
	return strings.HasPrefix(value, "$2a$") || strings.HasPrefix(value, "$2b$") || strings.HasPrefix(value, "$2y$")
}

func allowedRolesForClient(clientType string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(clientType)) {
	case domain.UserRoleViewer:
		return []string{domain.UserRoleViewer}, nil
	case domain.UserRoleAnchor:
		return []string{domain.UserRoleAnchor, domain.UserRoleAdmin}, nil
	default:
		return nil, ErrInvalidClientType
	}
}

func registrationRoleForClient(clientType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(clientType)) {
	case domain.UserRoleViewer:
		return domain.UserRoleViewer, nil
	case domain.UserRoleAnchor:
		return domain.UserRoleAnchor, nil
	default:
		return "", ErrInvalidClientType
	}
}

func hasAllowedRole(role string, allowedRoles ...string) bool {
	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}

func nextUserID(role string) string {
	switch role {
	case domain.UserRoleAnchor:
		return fmt.Sprintf("anchor-%d", time.Now().UnixNano())
	default:
		return fmt.Sprintf("user-%d", time.Now().UnixNano())
	}
}

func defaultAvatarForRole(role string) string {
	switch role {
	case domain.UserRoleAnchor:
		return "https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=200&q=80"
	default:
		return "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&q=80"
	}
}

func IsAdminRole(role string) bool {
	return role == domain.UserRoleAnchor || role == domain.UserRoleAdmin
}
