package fsmonitor

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// getFileChecksumAndSize calculates the SHA256 checksum and size of the file.
func getFileChecksumAndSize(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", 0, err
	}

	stat, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), stat.Size(), nil
}
