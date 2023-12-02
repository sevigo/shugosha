// Package fsmonitor provides a setup for monitoring file system changes.
package fsmonitor

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Monitor provides file system monitoring.
type Monitor struct {
	watcher      *fsnotify.Watcher
	eventHandler EventHandler
	mu           sync.Mutex
	eventBuffer  map[string][]fsnotify.Event
	bufferLock   sync.Mutex
	flushTimer   *time.Timer
	flushDelay   time.Duration
	subscribers  []Subscriber
	subLock      sync.Mutex // Protects the subscribers slice
}

// New creates a new Monitor instance.
func New(cfg *Config) (*Monitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Monitor{
		watcher:     watcher,
		eventBuffer: make(map[string][]fsnotify.Event), // Initialize the eventBuffer
		flushDelay:  cfg.FlushDelay,
		subscribers: make([]Subscriber, 0),
	}, nil
}

// Subscribe adds a new subscriber to the Monitor.
func (m *Monitor) Subscribe(sub Subscriber) {
	m.subLock.Lock()
	defer m.subLock.Unlock()

	m.subscribers = append(m.subscribers, sub)
}

// Unsubscribe removes a subscriber from the Monitor.
func (m *Monitor) Unsubscribe(sub Subscriber) {
	m.subLock.Lock()
	defer m.subLock.Unlock()

	for i, subscriber := range m.subscribers {
		if subscriber == sub {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			break
		}
	}
}

// Add adds a new directory to the watch list.
func (m *Monitor) Add(path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return m.watcher.Add(path)
		}

		if !info.IsDir() {
			m.handleEvent(fsnotify.Event{
				Name: path,
				Op:   fsnotify.Create, // Using Create as an equivalent for 'added'
			})
		}

		return nil
	})
}

// Remove removes a directory from the watch list.
func (m *Monitor) Remove(path string) error {
	return m.watcher.Remove(path)
}

// Start begins monitoring for file system events.
func (m *Monitor) Start() {
	m.eventBuffer = make(map[string][]fsnotify.Event) // Initialize buffer

	go func() {
		// Periodic flush timer
		flushTicker := time.NewTicker(m.flushDelay)
		defer flushTicker.Stop()

		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return
				}
				m.handleEvent(event)

			case err, ok := <-m.watcher.Errors:
				if !ok {
					return
				}
				slog.Error("got an error event from file watcher", "error", err)

			case <-flushTicker.C:
				m.flushEvents()
			}
		}
	}()
}

// Stop stops the monitor.
func (m *Monitor) Stop() error {
	if m.flushTimer != nil {
		m.flushTimer.Stop()
	}
	return m.watcher.Close()
}

// handleEvent processes the fsnotify events and buffers them.
func (m *Monitor) handleEvent(event fsnotify.Event) {
	m.bufferLock.Lock()
	defer m.bufferLock.Unlock()

	slog.Debug("[monitor] handle", "event", event.Op.String(), "file", event.Name)

	// Buffer the event
	m.eventBuffer[event.Name] = append(m.eventBuffer[event.Name], event)

	// Reset the flush timer
	if m.flushTimer != nil {
		m.flushTimer.Stop()
	}

	m.flushTimer = time.AfterFunc(m.flushDelay, m.flushEvents)
}

// flushEvents processes the buffered events to determine the final state.
func (m *Monitor) flushEvents() {
	m.bufferLock.Lock()
	defer m.bufferLock.Unlock()

	for _, events := range m.eventBuffer {
		finalEvent := determineFinalEvent(events)
		if finalEvent != nil {
			m.emitEvent(*finalEvent)
		}
	}

	// Clear the buffer after processing
	m.eventBuffer = make(map[string][]fsnotify.Event)
}

// emitEvent triggers the user-defined event handler.
func (m *Monitor) emitEvent(event Event) {
	m.subLock.Lock()
	defer m.subLock.Unlock()

	for _, sub := range m.subscribers {
		sub.HandleEvent(event)
	}
}
