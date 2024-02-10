package relay

import "sync"

type IPLimiter struct {
	Max    int
	mut    sync.Mutex
	active map[string]int
}

func NewIPLimiter(max int) *IPLimiter {
	return &IPLimiter{
		Max:    max,
		active: make(map[string]int),
	}
}

func (l *IPLimiter) Acquire(key string) bool {
	l.mut.Lock()
	defer l.mut.Unlock()

	if l.active[key] < l.Max {
		l.active[key]++
		return true
	}

	return false
}

func (l *IPLimiter) Release(key string) {
	l.mut.Lock()
	defer l.mut.Unlock()

	if _, ok := l.active[key]; ok {
		l.active[key]--

		if l.active[key] == 0 {
			delete(l.active, key)
		}
	}
}
