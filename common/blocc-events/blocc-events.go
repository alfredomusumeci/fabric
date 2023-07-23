package event

import (
	"sync"
)

// Event - BSCC Information to send a transaction successfully to the orderer
type Event struct {
	ChannelID   string
	SensoryTxID string
}

type Bus struct {
	subscribers []chan Event
	mu          sync.Mutex
}

func NewEventBus() *Bus {
	return &Bus{
		subscribers: []chan Event{},
		mu:          sync.Mutex{},
	}
}

// Subscribe - Subscribe to the event bus to receive events
func (bus *Bus) Subscribe() <-chan Event {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	ch := make(chan Event)
	bus.subscribers = append(bus.subscribers, ch)
	return ch
}

// Unsubscribe - Unsubscribe from the event bus
func (bus *Bus) Unsubscribe(ch <-chan Event) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	for i, subscriber := range bus.subscribers {
		if subscriber == ch {
			// Delete without preserving order
			bus.subscribers[i] = bus.subscribers[len(bus.subscribers)-1]
			bus.subscribers = bus.subscribers[:len(bus.subscribers)-1]
			break
		}
	}
}

// Publish - Publish an event to all subscribers
func (bus *Bus) Publish(event Event) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	for _, ch := range bus.subscribers {
		go func(ch chan Event) {
			ch <- event
		}(ch)
	}
}

var GlobalEventBus = NewEventBus()
