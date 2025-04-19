package service

import (
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/internal/publisher"
	"github.com/Koyo-os/Poll-service/internal/repository"
)

type PollService interface {
	Add(*entity.Poll) error
	Update(string, *entity.Poll) error
}

type PollServiceImpl struct {
	publisher  *publisher.Publisher
	repository repository.PollRepository
}

func Init(repository repository.PollRepository, pub *publisher.Publisher) PollService {
	return &PollServiceImpl{
		repository: repository,
		publisher:  pub,
	}
}

func (serviceImpl *PollServiceImpl) Add(poll *entity.Poll) error {
	if err := serviceImpl.repository.Add(poll); err != nil {
		return err
	} else {
		if err = serviceImpl.publisher.Publish(poll, "poll.created"); err != nil {
			return err
		}

		return nil
	}
}

func (serviceImpl *PollServiceImpl) Update(id string, poll *entity.Poll) error {
	if err := serviceImpl.repository.Update(id, poll); err != nil {
		return err
	} else {
		if err = serviceImpl.publisher.Publish(poll, "poll.updated"); err != nil {
			return err
		}

		return nil
	}
}
