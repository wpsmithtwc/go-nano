package handler

import (
	"errors"
	"time"

	"github.com/mouadino/go-nano/protocol"
	"github.com/mouadino/go-nano/transport"
)

// TODO: Trace-id Middleware.
// TODO: Log middleware.
// TODO: RateLimit middleware.

var (
	TimeOutError = errors.New("Timeout")
)

type TimeoutMiddleware struct {
	Handler
	timeout time.Duration
	fail    chan struct{}
	finish  chan error
}

func WithTimeout(timeout time.Duration) Middleware {
	return func(h Handler) Handler {
		return &TimeoutMiddleware{
			Handler: h,
			timeout: timeout,
			fail:    make(chan struct{}, 1),
			finish:  make(chan error, 1),
		}
	}
}

func (h *TimeoutMiddleware) Handle(w transport.ResponseWriter, r *protocol.Request) error {
	defer close(h.finish)
	defer close(h.fail)
	go func() {
		// TODO: Context and cancellation.
		h.finish <- h.Handler.Handle(w, r)
	}()
	go func() {
		time.Sleep(h.timeout * time.Second)
		h.fail <- struct{}{}
	}()

	select {
	case <-h.fail:
		return TimeOutError
	case err := <-h.finish:
		return err
	}
}
