package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
)

const userCacheTTLSeconds = 600

type userCache struct {
	client *realtime.Client
}

func newUserCache(client *realtime.Client) *userCache {
	if client == nil {
		return nil
	}
	return &userCache{client: client}
}

func (c *userCache) GetByID(userID string) (model.User, bool, error) {
	if c == nil {
		return model.User{}, false, nil
	}

	value, ok, err := c.client.Get(userCacheKeyByID(userID))
	if err != nil || !ok {
		return model.User{}, ok, err
	}

	var user model.User
	if err := json.Unmarshal([]byte(value), &user); err != nil {
		return model.User{}, false, err
	}
	return user, true, nil
}

func (c *userCache) Set(user model.User) error {
	if c == nil || strings.TrimSpace(user.ID) == "" {
		return nil
	}

	cachedUser := user
	cachedUser.Password = ""

	payload, err := json.Marshal(cachedUser)
	if err != nil {
		return err
	}
	return c.client.SetEX(userCacheKeyByID(user.ID), userCacheTTLSeconds, string(payload))
}

func (c *userCache) Delete(userID string) error {
	if c == nil || strings.TrimSpace(userID) == "" {
		return nil
	}
	return c.client.Del(userCacheKeyByID(userID))
}

func userCacheKeyByID(userID string) string {
	return fmt.Sprintf("user:profile:%s", userID)
}
