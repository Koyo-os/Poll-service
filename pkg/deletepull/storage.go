package deletepull

import (
	"context"
	"time"

	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const STORAGE_HASHMAP_KEY = "deletepull"

type (
	pullStorage struct {
		client *redis.Client
		logger *logger.Logger
	}

	Note struct {
		PollID  string        `json:"poll_id"`
		TimeEnd time.Duration `json:"time_end"`
	}
)

func initPullStorage(client *redis.Client, logger *logger.Logger) *pullStorage {
	return &pullStorage{
		client: client,
		logger: logger,
	}
}

func (store *pullStorage) CreateNote(ctx context.Context, value *Note) error {
	data, err := sonic.Marshal(value)
	if err != nil {
		store.logger.Error("error marshal note for",
			zap.String("poll_id", value.PollID),
			zap.Error(err))
		return err
	}

	uid := uuid.NewString()

	res := store.client.HSet(ctx, STORAGE_HASHMAP_KEY, uid, data)
	if err = res.Err(); err != nil {
		return err
	}

	return nil
}

func (storage *pullStorage) GetNotes(ctx context.Context) ([]Note, error) {
	result, err := storage.client.HGetAll(ctx, STORAGE_HASHMAP_KEY).Result()
	if err != nil {
		return nil, err
	}

	i := 0

	notes := make([]Note, 0, len(result))
	for _, data := range result {
		note := new(Note)

		if err = sonic.UnmarshalString(data, note); err != nil {
			storage.logger.Error("error unmarshal note", zap.Error(err))
			continue
		}

		notes[i] = *note
		i++
	}

	return notes, err
}
