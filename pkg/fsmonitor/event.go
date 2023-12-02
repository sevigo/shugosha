package fsmonitor

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event represents a file system event.
type Event struct {
	Path      string    // Path of the file/directory
	Type      string    // Type of event: "added", "changed", "deleted", "renamed"
	Timestamp time.Time // Time of the event
	Checksum  string    // SHA256 checksum of the file
	Size      int64     // Size of the file in bytes
}

// EventHandler is the function type for handling file system events.
type EventHandler func(Event)

// determineFinalEvent analyzes a slice of events and returns the final event.
func determineFinalEvent(events []fsnotify.Event) *Event {
	var created, changed, removed, renamed bool
	var lastEvent fsnotify.Event

	for _, event := range events {
		lastEvent = event

		switch {
		case event.Op&fsnotify.Create == fsnotify.Create:
			created = true
		case event.Op&fsnotify.Write == fsnotify.Write:
			changed = true
		case event.Op&fsnotify.Remove == fsnotify.Remove:
			removed = true
		case event.Op&fsnotify.Rename == fsnotify.Rename:
			renamed = true
		}
	}

	// Determine final event type and calculate checksum and size if needed
	var finalType string
	if created && !removed && !renamed {
		finalType = "added"
	} else if changed && !created && !renamed {
		finalType = "changed"
	} else {
		return nil
	}

	sum, size, _ := getFileChecksumAndSize(lastEvent.Name)
	return &Event{
		Path:      lastEvent.Name,
		Type:      finalType,
		Timestamp: time.Now(),
		Checksum:  sum,
		Size:      size,
	}
}
