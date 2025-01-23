package clock

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (r *RealClock) Now() time.Time {
	return time.Now()
}

func New() Clock {
	return &RealClock{}
}
