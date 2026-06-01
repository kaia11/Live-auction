package service

import (
	"time"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

type AuditService struct {
	logRepo repository.OperationLogRepository
}

func NewAuditService(logRepo repository.OperationLogRepository) *AuditService {
	return &AuditService{logRepo: logRepo}
}

func (s *AuditService) CreateLog(module, action, operatorID, roomID, targetType, targetID, content string) error {
	if s == nil || s.logRepo == nil {
		return nil
	}

	return s.logRepo.CreateLog(model.OperationLog{
		Module:     module,
		Action:     action,
		OperatorID: operatorID,
		RoomID:     roomID,
		TargetType: targetType,
		TargetID:   targetID,
		Content:    content,
		CreateTime: time.Now().Format(time.RFC3339),
	})
}
