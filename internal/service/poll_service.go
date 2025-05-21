package service

import (
	"context"
	"errors"
	"time"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/retrier"
	"github.com/google/uuid"
)

type PollServiceImpl struct {
	publisher  Publisher
	repository PollRepository
	waiter     Waiter
	casher     Casher
}

func Init(repository PollRepository, pub Publisher, casher Casher, waiter Waiter) *PollServiceImpl {
	return &PollServiceImpl{
		repository: repository,
		publisher:  pub,
		casher:     casher,
		waiter:     waiter,
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

		if poll.LimitedForTime {
			serviceImpl.waiter.AddDeleteWaiter(time.Until(poll.DeleteIn), poll.ID)
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

		if err := retrier.Do(3, 5, func() error {
			return serviceImpl.casher.DoCashing(context.Background(), poll.ID.String(), poll)
		}); err != nil {
			return err
		}

		if err = <-cherr; err != nil {
			return err
		}

		return nil
	}
}

func (serviceImpl *PollServiceImpl) SetPollClosed(pollId string) error {
	var cherr chan error

	uid, err := uuid.Parse(pollId)
	if err != nil {
		return err
	}

	poll, err := serviceImpl.repository.GetOne(uid)
	if err != nil {
		return err
	}

	if poll.Closed {
		return errors.New("poll is already closed")
	}

	poll.Closed = true

	if err = serviceImpl.repository.Update(uid, poll); err != nil {
		return err
	}

	go func() {
		cherr <- serviceImpl.publisher.Publish(poll, "poll.updated")
	}()

	if err = retrier.Do(3, 5, func() error {
		return serviceImpl.casher.DoCashing(context.Background(), poll.ID.String(), poll)
	}); err != nil {
		return err
	}
	if err = <-cherr; err != nil {
		return err
	}

	return nil
}
