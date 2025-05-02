package repository

import (
	"fmt"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (repoImpl *PollRepositoryImpl) GetOne(id uuid.UUID) (*entity.Poll, error) {
	var poll entity.Poll

	res := repoImpl.db.Where(&entity.Poll{
		ID: id,
	}).Find(&poll)

	if err := res.Error; err != nil {
		repoImpl.logger.Error("error get poll",
			zap.String("poll_id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error get poll: %v", err)
	}

	return &poll, nil
}
