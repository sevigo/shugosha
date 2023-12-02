package fsmonitor

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileChecksumAndSize(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "fsmonitor_test")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name()) // Clean up

	// Write some content to the file
	content := []byte("test content")
	_, err = tempFile.Write(content)
	assert.NoError(t, err)
	tempFile.Close() // Close the file to ensure the write is flushed

	// Expected checksum and size
	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256(content))
	expectedSize := int64(len(content))

	// Test getFileChecksumAndSize
	checksum, size, err := getFileChecksumAndSize(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedChecksum, checksum, "Checksum does not match")
	assert.Equal(t, expectedSize, size, "File size does not match")
}
