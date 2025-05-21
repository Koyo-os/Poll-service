package repository

import (
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (repo *PollRepositoryImpl) Delete(id uuid.UUID) error {
	res := repo.db.Delete(&entity.Poll{}, id)

	if err := res.Error; err != nil {
		repo.logger.Error("error delete poll from db",
			zap.String("poll_id", id.String()),
			zap.Error(err),
		)

		return err
	}

	return nil
}
