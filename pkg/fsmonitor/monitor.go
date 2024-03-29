// Package fsmonitor provides a setup for monitoring file system changes.
package fsmonitor

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sevigo/shugosha/pkg/model"
)

// Monitor provides file system monitoring.
type Monitor struct {
	dirs        map[string]int
	watcher     *fsnotify.Watcher
	eventBuffer map[string][]fsnotify.Event
	bufferLock  sync.Mutex
	flushTimer  *time.Timer
	flushDelay  time.Duration
	subscribers []model.Subscriber
	subLock     sync.Mutex // Protects the subscribers slice
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
		subscribers: make([]model.Subscriber, 0),
		dirs:        make(map[string]int),
	}, nil
}

func (m *Monitor) RootDirs() []string {
	dirs := []string{}
	for dir, index := range m.dirs {
		if index > 0 {
			dirs = append(dirs, dir)
		}
	}

	return dirs
}

// Subscribe adds a new subscriber to the Monitor.
func (m *Monitor) Subscribe(sub model.Subscriber) {
	m.subLock.Lock()
	defer m.subLock.Unlock()

	m.subscribers = append(m.subscribers, sub)
}

// Unsubscribe removes a subscriber from the Monitor.
func (m *Monitor) Unsubscribe(sub model.Subscriber) {
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
	_, isMonitored := m.dirs[path]
	if isMonitored {
		m.dirs[path]++
		return nil
	}

	m.dirs[path] = 1

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
	m.dirs[path]--
	return m.watcher.Remove(path)
}

// Start begins monitoring for file system events.
func (m *Monitor) Start(ctx context.Context) error {
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

			case <-ctx.Done():
				// Context is cancelled, perform cleanup and exit
				if err := m.Stop(); err != nil {
					slog.Error("Failed to stop monitor on context cancellation", "error", err)
				}
				return
			}
		}
	}()

	return nil
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
	if m.flushTimer == nil {
		m.flushTimer = time.AfterFunc(m.flushDelay, m.flushEvents)
	} else {
		m.flushTimer.Reset(m.flushDelay)
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
func (m *Monitor) emitEvent(event model.Event) {
	// Determine the root directory for the event
	for root := range m.dirs {
		if strings.HasPrefix(event.Path, root) {
			event.Root = root
			break
		}
	}

	m.subLock.Lock()
	for _, sub := range m.subscribers {
		sub.HandleEvent(event)
	}
	m.subLock.Unlock()
}
