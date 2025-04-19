package repository

import (
	"github.com/Koyo-os/Poll-service/internal/entity"
	"go.uber.org/zap/zapcore"
)

func (repoImpl *PollRepositoryImpl) Add(poll *entity.Poll) error {
	res := repoImpl.db.Create(poll)

	if err := res.Error; err != nil {
		repoImpl.logger.Error("error add poll", zapcore.Field{
			Key:    "err",
			String: err.Error(),
		})

		return err
	}

	repoImpl.logger.Info("successfully add poll with", zapcore.Field{
		Key:    "ID",
		String: poll.ID.String(),
	})

	return nil
}
