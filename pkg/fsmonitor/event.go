package fsmonitor

import (
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/sevigo/shugosha/pkg/model"
)

// determineFinalEvent analyzes a slice of events and returns the final event.
func determineFinalEvent(events []fsnotify.Event) *model.Event {
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
	return &model.Event{
		Path:      lastEvent.Name,
		Type:      finalType,
		Timestamp: time.Now(),
		Checksum:  sum,
		Size:      size,
	}
}
