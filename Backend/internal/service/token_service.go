package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type tokenClaims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Iat  int64  `json:"iat"`
	Exp  int64  `json:"exp"`
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string) *TokenService {
	if strings.TrimSpace(secret) == "" {
		secret = "replace_me"
	}

	return &TokenService{
		secret: []byte(secret),
		ttl:    24 * time.Hour,
	}
}

func (s *TokenService) Sign(userID, role string) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		Sub:  userID,
		Role: role,
		Iat:  now.Unix(),
		Exp:  now.Add(s.ttl).Unix(),
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerRaw, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsRaw, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerRaw)
	claimsPart := base64.RawURLEncoding.EncodeToString(claimsRaw)
	unsigned := headerPart + "." + claimsPart

	signature := signHMAC(unsigned, s.secret)
	return unsigned + "." + signature, nil
}

func (s *TokenService) Parse(authorization string) (tokenClaims, error) {
	token := strings.TrimSpace(authorization)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	if token == "" {
		return tokenClaims{}, ErrUnauthorizedToken
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, ErrUnauthorizedToken
	}

	unsigned := parts[0] + "." + parts[1]
	expected := signHMAC(unsigned, s.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return tokenClaims{}, ErrUnauthorizedToken
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, ErrUnauthorizedToken
	}

	var claims tokenClaims
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return tokenClaims{}, ErrUnauthorizedToken
	}

	if claims.Sub == "" || claims.Role == "" {
		return tokenClaims{}, ErrUnauthorizedToken
	}
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return tokenClaims{}, ErrUnauthorizedToken
	}

	return claims, nil
}

func signHMAC(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

var errInvalidToken = errors.New("invalid token")

func (s *TokenService) DebugDescribe(authorization string) string {
	claims, err := s.Parse(authorization)
	if err != nil {
		return errInvalidToken.Error()
	}
	return fmt.Sprintf("sub=%s role=%s exp=%d", claims.Sub, claims.Role, claims.Exp)
}
