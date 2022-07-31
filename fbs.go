package fbs

import (
	"errors"
	"github.com/google/uuid"
	"sync"
	"time"
)

type ChannelSingle interface {
	FeedBack() <-chan backInfo
	GetSingle() interface{}
	ToBack(val interface{})
}

type backInfo struct {
	backVal interface{}
	err     error
}

type channelSingle struct {
	source interface{}
	back   chan backInfo
	mu     sync.Mutex
	timer  *time.Timer
	in     bool
}

func (s *channelSingle) GetSingle() interface{} {
	return s.source
}

func (s *channelSingle) ToBack(val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.in {
		s.back <- backInfo{val, nil}
		s.timer.Stop()
		s.in = true
	}
}

func (s *channelSingle) FeedBack() <-chan backInfo {
	return s.back
}

func (s *channelSingle) timeout() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.in {
		s.back <- backInfo{nil, errors.New("timeout")}
		s.in = true
	}
}

// WithChannel 产生单个信号
func WithChannel(val interface{}, d time.Duration) ChannelSingle {
	if d <= 0 {
		panic(errors.New("must set overtime"))
	} else {
		c := make(chan backInfo, 1)
		s := &channelSingle{source: val, back: c}
		s.timer = time.AfterFunc(d, s.timeout)
		return s
	}
}

type EventIdSingle interface {
	AddSingle(string, time.Duration) <-chan string
	Done(string, interface{})
	GetVal(string) (interface{}, bool)
}

type eventIdSingle struct {
	backMap map[string]chan string
	backVal map[string]interface{}
	mu      sync.Mutex
}

// AddSingle 向管理器中添加一个信号, 接收信号若是""空，表示timeout。 若不为空，GetVal(反馈值）进行取值
func (e *eventIdSingle) AddSingle(eventId string, timeout time.Duration) <-chan string {
	ch := make(chan string, 1)
	e.mu.Lock()
	e.backMap[eventId] = ch
	e.mu.Unlock()
	time.AfterFunc(timeout, func() {
		if c, ok := e.backMap[eventId]; ok {
			e.mu.Lock()
			c <- ""
			delete(e.backMap, eventId)
			e.mu.Unlock()
		}
	})
	return ch
}

// Done 监测到返回值,添加注入信号管理器
func (e *eventIdSingle) Done(eventId string, backData interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if ch, ok := e.backMap[eventId]; ok {
		id := uuid.NewString()
		e.backVal[id] = backData
		ch <- id
		delete(e.backMap, eventId)
	}
}

func (e *eventIdSingle) GetVal(id string) (interface{}, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if val, ok := e.backVal[id]; ok {
		delete(e.backVal, id)
		return val, true
	}
	return nil, false
}

func WithEventId() EventIdSingle {
	return &eventIdSingle{backMap: make(map[string]chan string), backVal: make(map[string]interface{})}
}

func RandomEventId() string {
	return uuid.NewString()
}
