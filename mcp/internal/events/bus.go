// Package events provides a publish/subscribe event bus for dashboard SSE.
//
// Subscribers receive events on a buffered channel.  Slow subscribers
// that do not consume events in time will have messages dropped rather
// than blocking the publisher, ensuring the bus never stalls the VFS
// or command hot path.
package events

import (
	"encoding/json"
	"sync"
	"time"
)

// Type classifies the kind of dashboard event.
type Type string

const (
	TypeVFSUpdate       Type = "vfs_update"
	TypeAgentRegistered Type = "agent_registered"
	TypeAgentRemoved    Type = "agent_removed"
	TypeCommandResult   Type = "command_result"
)

// Event is a single message published to subscribers.
type Event struct {
	Type      Type      `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// JSON serialises the event.  Errors are silently swallowed (events
// are best-effort for the dashboard).
func (e Event) JSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// Subscriber is a handle returned by Bus.Subscribe.  Read from Ch to
// receive events.  When done, pass this to Bus.Unsubscribe.
type Subscriber struct {
	Ch   chan Event
	done chan struct{}
}

// Done returns a channel that is closed when the subscriber is unsubscribed.
func (s *Subscriber) Done() <-chan struct{} { return s.done }

// Bus is the central event dispatcher.
type Bus struct {
	mu   sync.RWMutex
	subs map[*Subscriber]struct{}
}

// NewBus creates a ready-to-use event bus.
func NewBus() *Bus {
	return &Bus{subs: make(map[*Subscriber]struct{})}
}

// Subscribe creates a new subscriber with a buffered channel.
func (b *Bus) Subscribe() *Subscriber {
	b.mu.Lock()
	defer b.mu.Unlock()
	sub := &Subscriber{
		Ch:   make(chan Event, 64),
		done: make(chan struct{}),
	}
	b.subs[sub] = struct{}{}
	return sub
}

// Unsubscribe removes a subscriber and closes its Done channel.
func (b *Bus) Unsubscribe(sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subs[sub]; ok {
		delete(b.subs, sub)
		close(sub.done)
	}
}

// Publish sends an event to all current subscribers.  If a subscriber's
// channel buffer is full the event is dropped for that subscriber.
func (b *Bus) Publish(eventType Type, data any) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	evt := Event{
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	for sub := range b.subs {
		select {
		case sub.Ch <- evt:
		default:
		}
	}
}

// SubscriberCount returns how many active subscribers exist.
func (b *Bus) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}
