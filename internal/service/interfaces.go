package service

import "github.com/Koyo-os/Poll-service/internal/entity"

type (
	PollRepository interface {
		Add(*entity.Poll) error
		Update(string, *entity.Poll) error
	}

	Publisher interface {
		Publish(any, string) error
	}
)
