package pull

import (
	"time"

	"github.com/google/uuid"
)

type DeletePullClient struct {
	reqChan chan Request
}

func InitClient(reqChan chan Request) *DeletePullClient {
	return &DeletePullClient{
		reqChan: reqChan,
	}
}

func (d *DeletePullClient) AddDeleteWaiter(wait time.Duration, id uuid.UUID) {
	d.reqChan <- Request{
		WaitTime:  wait,
		PurposeID: id,
	}
}
