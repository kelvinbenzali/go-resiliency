package breaker

import (
	"sync/atomic"
	"time"
)

// Run will either return ErrBreakerOpen immediately if the circuit-breaker is
// already open, or it will run the given function and pass along its return
// value. It is safe to call Run concurrently on the same Breaker.
func (b *Breaker) Run(work func() error) error {
	state := atomic.LoadUint32(&b.state)

	if state == open {
		return ErrBreakerOpen
	}

	return b.doWork(state, work)
}

// Go will either return ErrBreakerOpen immediately if the circuit-breaker is
// already open, or it will run the given function in a separate goroutine.
// If the function is run, Go will return nil immediately, and will *not* return
// the return value of the function. It is safe to call Go concurrently on the
// same Breaker.
func (b *Breaker) Go(work func() error) error {
	state := atomic.LoadUint32(&b.state)

	if state == open {
		return ErrBreakerOpen
	}

	// errcheck complains about ignoring the error return value, but
	// that's on purpose; if you want an error from a goroutine you have to
	// get it over a channel or something
	go b.doWork(state, work)

	return nil
}

func (b *Breaker) doWork(state uint32, work func() error) error {
	var panicValue interface{}

	result := func() error {
		defer func() {
			panicValue = recover()
		}()
		return work()
	}()

	if result == nil && panicValue == nil && state == closed {
		// short-circuit the normal, success path without contending
		// on the lock
		return nil
	}

	// oh well, I guess we have to contend on the lock
	b.processResult(result, panicValue)

	if panicValue != nil {
		// as close as Go lets us come to a "rethrow" although unfortunately
		// we lose the original panicing location
		panic(panicValue)
	}

	return result
}

func (b *Breaker) processResult(result error, panicValue interface{}) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if result == nil && panicValue == nil && b.state == halfOpen {
		b.successes++
		if b.successes == b.config.SuccessThreshold {
			b.closeBreaker()
		}

	} else {

		switch b.state {
		case closed:
			b.lastError = b.removeExpiredError()
			b.lastError = append(b.lastError, time.Now())
			if len(b.lastError) == b.config.ErrorThreshold {
				b.openBreaker()
			}
		case halfOpen:
			b.openBreaker()
		}
	}
}

func (b *Breaker) openBreaker() {
	b.changeState(open)
	go b.timer()
}

func (b *Breaker) closeBreaker() {
	b.changeState(closed)
}

func (b *Breaker) timer() {
	time.Sleep(b.config.TimeoutOpen)

	b.lock.Lock()
	defer b.lock.Unlock()

	b.changeState(halfOpen)
}

func (b *Breaker) changeState(newState uint32) {
	b.lastError = []time.Time{}
	b.successes = 0
	atomic.StoreUint32(&b.state, newState)
}

func (b *Breaker) removeExpiredError() []time.Time {
	var newErrorList []time.Time
	for _, v := range b.lastError {
		if v.After(time.Now().Add(-b.config.TimeoutClosed)) {
			newErrorList = append(newErrorList, v)
		}
	}
	return newErrorList
}
