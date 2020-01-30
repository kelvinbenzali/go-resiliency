circuit-breaker
===============

[![Build Status](https://travis-ci.org/eapache/go-resiliency.svg?branch=master)](https://travis-ci.org/eapache/go-resiliency)
[![GoDoc](https://godoc.org/github.com/eapache/go-resiliency/breaker?status.svg)](https://godoc.org/github.com/eapache/go-resiliency/breaker)
[![Code of Conduct](https://img.shields.io/badge/code%20of%20conduct-active-blue.svg)](https://eapache.github.io/conduct.html)

The circuit-breaker resiliency pattern for golang.

Creating a breaker takes four parameters:
- error threshold (for opening the breaker after error reached threshold within the timeout closed window time)
- success threshold (for closing the breaker after consecutive number of success threshold reached)
- timeout closed (window time after the first error detected when breaker closed)
- timeout open (how long to keep the breaker open before changing state to half-open)

```go
b := breaker.New(
    BreakerConfig{
    	ErrorThreshold:     3,
        SuccessThreshold:   1, 
        TimeoutClosed:      5*time.Second,
        TimeoutOpen:        5*time.Second,
    }
)

for {
	result := b.Run(func() error {
		// communicate with some external service and
		// return an error if the communication failed
		return nil
	})

	switch result {
	case nil:
		// success!
	case breaker.ErrBreakerOpen:
		// our function wasn't run because the breaker was open
	default:
		// some other error
	}
}
```
