package service

import (
	"context"
	"time"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/google/uuid"
)

type (
	PollRepository interface {
		Add(*entity.Poll) error
		GetOne(uuid.UUID) (*entity.Poll, error)
		Update(uuid.UUID, *entity.Poll) error
	}

	Publisher interface {
		Publish(any, string) error
	}

	Casher interface {
		DoCashing(ctx context.Context, key string, payload any) error // payload must to be pointer
	}

	Waiter interface {
		AddDeleteWaiter(time.Duration, uuid.UUID)
	}
)
