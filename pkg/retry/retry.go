package retry

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type Retrier struct {
	baseTimeout time.Duration
	attemps     int
}

var defaultRetrier = &Retrier{
	baseTimeout: time.Duration(1) * time.Second,
	attemps:     3,
}

func Default() *Retrier {
	return defaultRetrier
}

var sleep = time.Sleep

type retriableFn func() error

func IfTemporary(ctx context.Context, fn retriableFn) error {
	return defaultRetrier.IfTemporary(ctx, fn)
}

func (r *Retrier) IfTemporary(ctx context.Context, fn retriableFn) error {
	var err error
	for attempt := 1; ; attempt++ {
		err = fn()
		if !isTemporary(err) || attempt >= r.attemps {
			break
		}
		sleep(time.Duration(int64(attempt) * int64(r.baseTimeout)))
	}
	return err
}

func isTemporary(err error) bool {
	e, ok := errors.Cause(err).(interface {
		Temporary() bool
	})
	return ok && e.Temporary()
}
