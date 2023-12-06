package model

import "time"

// Subscriber defines the interface for event handlers.
type Subscriber interface {
	HandleEvent(Event)
}

// Event represents a file system event.
type Event struct {
	Root      string
	Path      string    // Path of the file/directory
	Type      string    // Type of event: "added", "changed", "deleted", "renamed"
	Timestamp time.Time // Time of the event
	Checksum  string    // SHA256 checksum of the file
	Size      int64     // Size of the file in bytes
}
