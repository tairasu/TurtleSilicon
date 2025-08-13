package epochsilicon

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"turtlesilicon/pkg/debug"
)

// calculateFileHash calculates the MD5 hash of a local file
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getRemoteFileMetadata gets the metadata (size and last-modified) of a remote file
func getRemoteFileMetadata(url string) (*FileMetadata, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d for %s", resp.StatusCode, url)
	}

	metadata := &FileMetadata{}

	// Get file size
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			metadata.Size = size
		}
	}

	// Get last modified date
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if modTime, err := time.Parse(time.RFC1123, lastModified); err == nil {
			metadata.LastModified = modTime
		}
	}

	return metadata, nil
}

// getLocalFileMetadata gets the metadata of a local file
func getLocalFileMetadata(filePath string) (*FileMetadata, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &FileMetadata{
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
	}, nil
}

// verifyFileHash verifies that a local file matches the expected hash
func verifyFileHash(filePath, expectedHash string) error {
	if expectedHash == "" {
		debug.Printf("No hash provided for verification of %s", filePath)
		return nil
	}

	localHash, err := calculateFileHash(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %v", err)
	}

	if localHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected=%s, got=%s", expectedHash, localHash)
	}

	debug.Printf("Hash verification successful for %s: %s", filePath, localHash)
	return nil
}
