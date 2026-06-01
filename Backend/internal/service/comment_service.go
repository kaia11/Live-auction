package service

import (
	"time"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

type CommentService struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) CreateRoomComment(roomID, userID, nickname, content string) (model.RoomComment, error) {
	comment := model.RoomComment{
		RoomID:     roomID,
		UserID:     userID,
		Nickname:   nickname,
		Content:    content,
		CreateTime: time.Now().Format(time.RFC3339),
	}

	if s.repo != nil {
		if err := s.repo.CreateComment(comment); err != nil {
			return model.RoomComment{}, err
		}
	}

	return comment, nil
}
