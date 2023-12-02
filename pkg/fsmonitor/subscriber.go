package fsmonitor

// Subscriber defines the interface for event handlers.
type Subscriber interface {
	HandleEvent(Event)
}
