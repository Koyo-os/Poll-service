package repository

import (
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (repoImpl *PollRepositoryImpl) Update(uid uuid.UUID, poll *entity.Poll) error {
	res := repoImpl.db.Where(&entity.Poll{
		ID: uid,
	},
	).Updates(poll)

	if err := res.Error; err != nil {
		repoImpl.logger.Error("error update poll with",
			zap.String("poll_id", uid.String()),
			zap.Error(err),
		)

		return err
	}
	repoImpl.logger.Info("successfully update poll with",
		zap.String("poll_id", uid.String()),
	)

	return nil
}
