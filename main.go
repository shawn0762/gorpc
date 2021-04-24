package main

import (
	"errors"
	"sync"
)

func main() {
	t := &Test{}
	_, _ = t.Test()
}

type Test struct {
	mu sync.Mutex
}

func (t *Test) isTrue() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return true
}

func (t *Test) Test() (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.isTrue() {
		return false, errors.New("something went wrong")
	}
	return true, nil
}
