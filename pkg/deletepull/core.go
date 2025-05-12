package deletepull

import (
	"context"
	"sync"
	"time"

	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

const MAX_ERR_COUNT = 20

var errCount int

type (
	Deleter interface {
		Delete(uuid.UUID) error
	}

	Request struct {
		PurposeID uuid.UUID
		WaitTime  time.Duration
	}

	DeletePullCore struct {
		eg      *errgroup.Group
		egCtx   context.Context
		mutex   *sync.Mutex
		reqChan chan Request
		errChan chan error
		deleter Deleter
		storage *pullStorage
	}
)

func Init(cherr chan error, client *redis.Client, logger *logger.Logger) *DeletePullCore {
	var mutex sync.Mutex

	eg, ctx := errgroup.WithContext(context.Background())

	return &DeletePullCore{
		mutex:   &mutex,
		eg:      eg,
		egCtx:   ctx,
		errChan: cherr,
		storage: initPullStorage(client, logger),
	}
}

func (d *DeletePullCore) BeforeRun() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	notes, err := d.storage.GetNotes(ctx)
	if err != nil {
		return err
	}

	for _, n := range notes {
		uid, err := uuid.Parse(n.PollID)
		if err != nil {
			continue
		}

		d.reqChan <- Request{
			WaitTime:  n.TimeEnd,
			PurposeID: uid,
		}
	}

	return nil
}

func (d *DeletePullCore) routeError(err error) error {
	d.mutex.Lock()
	errCount++

	if errCount >= MAX_ERR_COUNT {
		d.mutex.Unlock()
		return context.Canceled
	}
	d.mutex.Unlock()
	d.errChan <- err

	return nil
}

func (d *DeletePullCore) wait(req *Request) error {
	select {
	case <-time.After(req.WaitTime):
		if err := d.storage.CreateNote(context.Background(), &Note{
			TimeEnd: req.WaitTime,
			PollID:  req.PurposeID.String(),
		}); err != nil {
			return d.routeError(err)
		}
		if err := d.deleter.Delete(req.PurposeID); err != nil {
			return d.routeError(err)
		}
		return nil
	case <-d.egCtx.Done():
		return d.egCtx.Err()
	}
}

func (d *DeletePullCore) Listen(ctx context.Context) {
	for {
		select {
		case req := <-d.reqChan:
			d.eg.Go(func() error {
				return d.wait(&req)
			})
		case <-d.egCtx.Done():
			return
		case <-ctx.Done():
			return
		}
	}
}
