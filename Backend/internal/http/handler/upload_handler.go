package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/service"
)

const maxUploadSize = 5 << 20

type UploadHandler struct {
	uploadDir   string
	userService *service.UserService
}

type uploadImageResponse struct {
	URL string `json:"url"`
}

func NewUploadHandler(uploadDir string, userService *service.UserService) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir, userService: userService}
}

func (h *UploadHandler) UploadImage(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	r.Body = nethttp.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		api.BadRequest(w, "image must be 5MB or smaller")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file is required")
		return
	}
	defer file.Close()

	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && !errors.Is(err, io.EOF) {
		api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to read upload")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to read upload")
		return
	}

	contentType := nethttp.DetectContentType(sniff[:n])
	ext, ok := extensionForImage(contentType)
	if !ok {
		api.BadRequest(w, "only jpeg, png, gif and webp images are supported")
		return
	}

	filename, err := buildUploadFilename(ext)
	if err != nil {
		api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to create upload filename")
		return
	}

	targetPath := filepath.Join(h.uploadDir, filename)
	out, err := os.Create(targetPath)
	if err != nil {
		api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to save upload")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		_ = os.Remove(targetPath)
		api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to save upload")
		return
	}

	_ = header
	api.Created(w, uploadImageResponse{URL: "/uploads/" + filename})
}

func (h *UploadHandler) ServeUpload(w nethttp.ResponseWriter, r *nethttp.Request) {
	name := r.PathValue("file")
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `\`) {
		nethttp.NotFound(w, r)
		return
	}

	targetPath := filepath.Join(h.uploadDir, filepath.Clean(name))
	cleanUploadDir := filepath.Clean(h.uploadDir)
	if !strings.HasPrefix(targetPath, cleanUploadDir+string(os.PathSeparator)) && targetPath != cleanUploadDir {
		nethttp.NotFound(w, r)
		return
	}

	nethttp.ServeFile(w, r, targetPath)
}

func (h *UploadHandler) ensureAdminAccess(w nethttp.ResponseWriter, r *nethttp.Request) bool {
	_, err := h.userService.RequireAnyRole(r.Header.Get("Authorization"), domain.UserRoleAnchor, domain.UserRoleAdmin)
	if err == nil {
		return true
	}

	switch {
	case errors.Is(err, service.ErrUnauthorizedToken):
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbiddenRole):
		api.Error(w, nethttp.StatusForbidden, api.CodeForbidden, err.Error())
	default:
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
	}

	return false
}

func extensionForImage(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func buildUploadFilename(ext string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes), ext), nil
}
