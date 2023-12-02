package fsmonitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// echoSubscriber is a simple subscriber that sends them back to a channel for testing.
type echoSubscriber struct {
	EventChannel chan Event
}

// HandleEvent is called when a file system event is received.
func (s *echoSubscriber) HandleEvent(event Event) {
	// Send the event to the channel for testing
	if s.EventChannel != nil {
		s.EventChannel <- event
	}
}

// setupTestEnvironment creates a temporary directory with a test file.
func setupTestEnvironment(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "fsmonitor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	tempFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(tempFile, []byte("test content"), 0666); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir) // clean up
	}

	return tempDir, cleanup
}

// TestMonitorEvent implementation
func TestMonitorEvent(t *testing.T) {
	// Setup environment
	watchedDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	monitor, err := New(&Config{FlushDelay: 500 * time.Millisecond})
	assert.NoError(t, err, "Failed to create monitor")
	defer monitor.Stop()

	// Custom subscriber to record events
	events := make(chan Event, 10) // Buffered channel
	subscriber := &echoSubscriber{EventChannel: events}
	monitor.Subscribe(subscriber)

	// Add directory to watch
	assert.NoError(t, monitor.Add(watchedDir), "Failed to add directory to monitor")

	// Start monitoring
	go monitor.Start()

	// Perform a file operation to trigger an event
	newFilePath := filepath.Join(watchedDir, "newfile.txt")
	assert.NoError(t, os.WriteFile(newFilePath, []byte("new file content"), 0666), "Failed to create a new file")

	// Wait for the event or timeout
	select {
	case event := <-events:
		assert.Equal(t, "added", event.Type, "Event type mismatch")
		assert.Equal(t, newFilePath, event.Path, "Event path mismatch")

	case <-time.After(1000 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

func TestMonitorRemove(t *testing.T) {
	watchedDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	monitor, err := New(&Config{FlushDelay: 100 * time.Millisecond})
	assert.NoError(t, err)
	defer monitor.Stop()

	events := make(chan Event, 10)
	subscriber := &echoSubscriber{EventChannel: events}
	monitor.Subscribe(subscriber)

	assert.NoError(t, monitor.Add(watchedDir))
	go monitor.Start()

	// Create a new file to trigger an event
	testFilePath := filepath.Join(watchedDir, "remove_test.txt")
	assert.NoError(t, os.WriteFile(testFilePath, []byte("test"), 0666))

	// Wait for the event or timeout
	select {
	case <-events:
		// Expected to receive an event
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event before remove")
	}

	// Remove the directory from the watch list
	assert.NoError(t, monitor.Remove(watchedDir))

	// Perform another file operation
	assert.NoError(t, os.WriteFile(testFilePath, []byte("test update"), 0666))

	// Check that no more events are received
	select {
	case <-events:
		t.Error("Event received after directory was removed")
	case <-time.After(1 * time.Second):
		// Expected to not receive any event
	}
}

func TestMonitorUnsubscribe(t *testing.T) {
	watchedDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	monitor, err := New(&Config{FlushDelay: 100 * time.Millisecond})
	assert.NoError(t, err)
	defer monitor.Stop()

	events := make(chan Event, 10)
	subscriber := &echoSubscriber{EventChannel: events}
	monitor.Subscribe(subscriber)

	assert.NoError(t, monitor.Add(watchedDir))
	go monitor.Start()

	// Create a new file to trigger an event
	testFilePath := filepath.Join(watchedDir, "unsubscribe_test.txt")
	assert.NoError(t, os.WriteFile(testFilePath, []byte("test"), 0666))

	// Wait for the event or timeout
	select {
	case <-events:
		// Expected to receive an event
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event before unsubscribe")
	}

	// Unsubscribe the subscriber
	monitor.Unsubscribe(subscriber)

	// Perform another file operation
	assert.NoError(t, os.WriteFile(testFilePath, []byte("test update"), 0666))

	// Check that no more events are received
	select {
	case <-events:
		t.Error("Event received after unsubscribe")
	case <-time.After(1 * time.Second):
		// Expected to not receive any event
	}
}
