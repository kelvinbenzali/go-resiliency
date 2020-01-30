package breaker

import (
	"sync"
	"time"
)

// Breaker implements the circuit-breaker resiliency pattern
type Breaker struct {
	config BreakerConfig

	lock      sync.Mutex
	state     uint32
	successes int
	lastError []time.Time
}

type BreakerConfig struct {
	ErrorThreshold   int
	SuccessThreshold int
	TimeoutClosed    time.Duration
	TimeoutOpen      time.Duration
}

// New constructs a new circuit-breaker that starts closed.
// From closed, the breaker opens if "errorThreshold" errors are seen
// without an error-free period of at least "timeout". From open, the
// breaker half-closes after "timeout". From half-open, the breaker closes
// after "successThreshold" consecutive successes, or opens on a single error.
func New(config BreakerConfig) *Breaker {
	return &Breaker{
		config: config,
	}
}
