package promise

import (
	"time"
)

// Promise the promise contract
type Promise interface {
	Until(func(interface{}), time.Duration) Promise
	Then(func(interface{})) Promise
	Fail(func(error)) Promise
	WhenCancelled(func()) Promise
	IsCancelled() bool
	Cancel()
}

type promiseImpl struct {
	cancel      chan struct{}
	done        chan struct{}
	val         interface{}
	res         func(func(interface{}))
	err         chan error
	isCancelled bool
}

// IsCancelled returns true if the promise has been cancelled, false otherwise
func (p *promiseImpl) IsCancelled() bool {
	return p.isCancelled
}

// WhenCancelled lets you handle promises that have been cancelled
func (p *promiseImpl) WhenCancelled(c func()) Promise {
	go func() {
		if p.IsCancelled() == true {
			c()
			return
		}
		select {
		case <-p.cancel:
			c()
		}
	}()
	return p
}

// Cacel cancels the promise if not already resolved
func (p *promiseImpl) Cancel() {
	if p.IsCancelled() {
		return
	}
	go func() {
		p.cancel <- struct{}{}
	}()

}

// Until is similar to Then except the promise will be cancelled afte a given time
func (p *promiseImpl) Until(then func(interface{}), to time.Duration) Promise {
	go func() {
		select {
		case <-p.done:
			then(p.val)
		case <-time.After(to):
			p.Cancel()
		}
	}()

	return p
}

// Then lets you specify the general handler when the promise is resolved
func (p *promiseImpl) Then(then func(interface{})) Promise {
	go func() {
		if p.IsCancelled() {
			return
		}

		select {
		case <-p.done:
			then(p.val)
		case <-p.cancel:
		}
	}()

	return p
}

// Fail lets you handle a failed promise
func (p *promiseImpl) Fail(handle func(error)) Promise {
	go func() {
		if p.IsCancelled() == true {
			return
		}
		select {
		case err := <-p.err:
			handle(err)
		}
	}()

	return p
}

func newInner(cancelChan chan struct{}, f func(func(interface{}), func(error))) Promise {
	inner := promiseImpl{
		done:        make(chan struct{}),
		cancel:      cancelChan,
		err:         make(chan error, 1),
		isCancelled: false,
	}

	res := func(v interface{}) {

		if inner.isCancelled {
			return
		}

		inner.val = v
		inner.done <- struct{}{}
	}

	ef := func(e error) {
		inner.err <- e
	}

	go func() {
		select {
		case <-inner.cancel:
			inner.isCancelled = true
		default:
			f(res, ef)
		}

	}()

	return &inner
}

// New creates a new Promise
func New(f func(func(interface{}), func(error))) Promise {
	c := make(chan struct{})
	p := newInner(c, f)

	return p
}
