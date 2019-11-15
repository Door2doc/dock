package history

import (
	"sync"
	"time"
)

const MaxHistory = 10

// History keeps track of recent events.
type History struct {
	mu     sync.Mutex
	events []*Event
}

func New() *History {
	return &History{}
}

// Event is a single line in the history.
type Event struct {
	Time           time.Time
	QueryDuration  time.Duration
	UploadDuration time.Duration
	Size           int
	JSON           string
	Error          error
}

func (h *History) NewEvent() *Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	e := &Event{Time: time.Now()}

	h.events = append([]*Event{e}, h.events...)
	if len(h.events) > MaxHistory {
		h.events = h.events[:MaxHistory]
	}

	return e
}

func (h *History) Events() []*Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	res := make([]*Event, len(h.events))
	copy(res, h.events)
	return res
}
