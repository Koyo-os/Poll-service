package repository

import (
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
)

func parseUUID(input string) (uuid.UUID, error) {
	return uuid.Parse(input)
}

func (repoImpl *PollRepositoryImpl) Update(uuid string, poll *entity.Poll) error {
	uid, err := parseUUID(uuid)
	if err != nil {
		return err
	}

	res := repoImpl.db.Where(&entity.Poll{
		ID: uid,
	},
	).Updates(poll)

	if err = res.Error; err != nil {
		repoImpl.logger.Error("error update poll with", zapcore.Field{
			Key:    "err",
			String: uuid,
		},
			zapcore.Field{
				Key:    "ID",
				String: uuid,
			})
	}

	repoImpl.logger.Info("successfully update poll with", zapcore.Field{
		Key:    "ID",
		String: uuid,
	})

	return nil
}
