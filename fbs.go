package fbs

import (
	"errors"
	"sync"
	"time"
)

type Single interface {
	FeedBack() <-chan backInfo
	GetSingle() interface{}
	ToBack(val interface{})
}

type backInfo struct {
	backVal interface{}
	err     error
}

type single struct {
	source interface{}
	back   chan backInfo
	mu     sync.Mutex
	timer  *time.Timer
	in     bool
}

func (s *single) GetSingle() interface{} {
	return s.source
}

func (s *single) ToBack(val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.in {
		s.back <- backInfo{val, nil}
		s.timer.Stop()
		s.in = true
	}
}

func (s *single) FeedBack() <-chan backInfo {
	return s.back
}

func (s *single) timeout() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.in {
		s.back <- backInfo{nil, errors.New("timeout")}
		s.in = true
	}
}

func NewSingle(val interface{}, d time.Duration) Single {
	if d <= 0 {
		panic(errors.New("must set overtime"))
	} else {
		c := make(chan backInfo, 1)
		s := &single{source: val, back: c}
		s.timer = time.AfterFunc(d, s.timeout)
		return s
	}
}
