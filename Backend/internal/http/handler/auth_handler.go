package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/service"
)

type AuthHandler struct {
	userService *service.UserService
}

type loginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ClientType string `json:"clientType"`
}

type registerRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ClientType string `json:"clientType"`
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) Login(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.userService.Login(req.Username, req.Password, req.ClientType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
		case errors.Is(err, service.ErrInvalidClientType):
			api.BadRequest(w, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeNotFound, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to login")
		}
		return
	}

	api.Success(w, nethttp.StatusOK, result)
}

func (h *AuthHandler) Register(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.userService.Register(req.Username, req.Password, req.ClientType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials), errors.Is(err, service.ErrInvalidClientType):
			api.BadRequest(w, err.Error())
		case errors.Is(err, service.ErrUsernameTaken):
			api.Error(w, nethttp.StatusConflict, api.CodeConflict, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to register")
		}
		return
	}

	api.Success(w, nethttp.StatusCreated, result)
}

func (h *AuthHandler) GetCurrentUser(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, err := h.userService.GetCurrentUser(r.Header.Get("Authorization"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorizedToken):
			api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeNotFound, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to get current user")
		}
		return
	}

	api.Success(w, nethttp.StatusOK, user)
}
