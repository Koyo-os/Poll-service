package service

import (
	"context"
	"errors"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/retrier"
	"github.com/google/uuid"
)

type PollService interface {
	Add(*entity.Poll) error
	Update(string, *entity.Poll) error
}

type PollServiceImpl struct {
	publisher  Publisher
	repository PollRepository
	casher     Casher
}

func Init(repository PollRepository, pub Publisher, casher Casher) PollService {
	return &PollServiceImpl{
		repository: repository,
		publisher:  pub,
		casher:     casher,
	}
}

func (serviceImpl *PollServiceImpl) Add(poll *entity.Poll) error {
	cherr := make(chan error, 1)

	if err := serviceImpl.repository.Add(poll); err != nil {
		return err
	} else {
		go func() {
			cherr <- serviceImpl.publisher.Publish(poll, "poll.created")
		}()

		err = retrier.Do(3, 5, func() error {
			return serviceImpl.casher.DoCashing(context.Background(), poll.ID.String(), poll)
		})
		if err != nil {
			return err
		}

		if err = <-cherr; err != nil {
			return err
		}

		return nil
	}
}

func (serviceImpl *PollServiceImpl) Update(id string, poll *entity.Poll) error {
	cherr := make(chan error, 1)

	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	if err := serviceImpl.repository.Update(uid, poll); err != nil {
		return err
	} else {
		go func() {
			cherr <- serviceImpl.publisher.Publish(poll, "poll.updated")
		}()

		err = retrier.Do(3, 5, func() error {
			return serviceImpl.casher.DoCashing(context.Background(), poll.ID.String(), poll)
		})
		if err != nil {
			return err
		}

		if err = <-cherr; err != nil {
			return err
		}

		return nil
	}
}

func (serviceImpl *PollServiceImpl) SetPollClosed(pollId string) error {
	uid, err := uuid.Parse(pollId)
	if err != nil {
		return err
	}

	poll, err := serviceImpl.repository.GetOne(uid)
	if err != nil {
		return err
	}

	if poll.Closed == true {
		return errors.New("poll is already closed")
	}

	poll.Closed = true

	return serviceImpl.repository.Update(uid, poll)
}
