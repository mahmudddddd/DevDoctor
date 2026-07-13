package runner

import "time"

type processController interface {
	Terminate(time.Duration) error
	Close() error
}
