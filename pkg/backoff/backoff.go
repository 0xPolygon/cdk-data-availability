package backoff

import "time"

// Exponential performs exponential backoff attempts on a given action
func Exponential(action func() error, attempts uint, wait time.Duration) error {
	var err error
	for i := uint(0); i < attempts; i++ {
		if err = action(); err == nil {
			return nil
		}
		time.Sleep(wait)
		wait *= 2
	}
	return err
}
