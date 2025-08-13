package epochsilicon

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"turtlesilicon/pkg/debug"
)

// ProgressWriter wraps an io.Writer and reports progress
type ProgressWriter struct {
	writer     io.Writer
	total      int64
	downloaded int64
	onProgress func(downloaded, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.downloaded += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.downloaded, pw.total)
	}
	return
}

// downloadFile downloads a single file to the correct location (legacy method)
func downloadFile(gamePath string, file RequiredFile) error {
	fullPath := filepath.Join(gamePath, file.RelativePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Download file
	resp, err := http.Get(file.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %v", file.DownloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for %s", resp.StatusCode, file.DownloadURL)
	}

	// Create the file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", fullPath, err)
	}
	defer outFile.Close()

	// Copy data
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", fullPath, err)
	}

	debug.Printf("Successfully downloaded: %s", file.DisplayName)
	return nil
}

// downloadFileWithProgress downloads a file and reports progress (legacy method)
func downloadFileWithProgress(gamePath string, file RequiredFile, onProgress func(downloaded, total int64)) error {
	fullPath := filepath.Join(gamePath, file.RelativePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Download file
	resp, err := http.Get(file.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %v", file.DownloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for %s", resp.StatusCode, file.DownloadURL)
	}

	// Get content length for progress tracking
	contentLength := resp.ContentLength

	// Create the file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", fullPath, err)
	}
	defer outFile.Close()

	// Create progress writer
	progressWriter := &ProgressWriter{
		writer:     outFile,
		total:      contentLength,
		onProgress: onProgress,
	}

	// Copy data with progress tracking
	_, err = io.Copy(progressWriter, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", fullPath, err)
	}

	debug.Printf("Successfully downloaded: %s", file.DisplayName)
	return nil
}

// downloadSingleFile downloads from a single URL
func downloadSingleFile(fullPath, url string, expectedSize int64, onProgress func(downloaded, total int64)) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for %s", resp.StatusCode, url)
	}

	// Use expected size if content length is not available
	contentLength := resp.ContentLength
	if contentLength <= 0 && expectedSize > 0 {
		contentLength = expectedSize
	}

	// Create the file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", fullPath, err)
	}
	defer outFile.Close()

	// Create progress writer if callback provided
	if onProgress != nil {
		progressWriter := &ProgressWriter{
			writer:     outFile,
			total:      contentLength,
			onProgress: onProgress,
		}
		_, err = io.Copy(progressWriter, resp.Body)
	} else {
		_, err = io.Copy(outFile, resp.Body)
	}

	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", fullPath, err)
	}

	return nil
}

// downloadFileWithVerification downloads a file and verifies its hash
func downloadFileWithVerification(gamePath string, file RequiredFile, onProgress func(downloaded, total int64)) error {
	fullPath := filepath.Join(gamePath, file.RelativePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Try CDNs in order of preference until one succeeds
	cdnPriority := strings.Split(CDNPriority, ",")
	var lastErr error

	for _, cdn := range cdnPriority {
		if url, exists := file.URLs[cdn]; exists {
			debug.Printf("Attempting download from %s CDN: %s", cdn, url)

			if err := downloadSingleFile(fullPath, url, file.Size, onProgress); err != nil {
				debug.Printf("Failed to download from %s CDN: %v", cdn, err)
				lastErr = err
				continue
			}

			// Verify hash if provided
			if err := verifyFileHash(fullPath, file.Hash); err != nil {
				debug.Printf("Hash verification failed for %s from %s CDN: %v", file.RelativePath, cdn, err)
				os.Remove(fullPath) // Remove corrupted file
				lastErr = err
				continue
			}

			debug.Printf("Successfully downloaded and verified: %s", file.DisplayName)
			return nil
		}
	}

	// If all CDNs failed, try the primary download URL as fallback
	if file.DownloadURL != "" {
		debug.Printf("Trying fallback URL: %s", file.DownloadURL)
		if err := downloadSingleFile(fullPath, file.DownloadURL, file.Size, onProgress); err != nil {
			return fmt.Errorf("all download attempts failed, last error: %v", err)
		}

		// Verify hash if provided
		if err := verifyFileHash(fullPath, file.Hash); err != nil {
			os.Remove(fullPath)
			return fmt.Errorf("hash verification failed for fallback download: %v", err)
		}

		debug.Printf("Successfully downloaded and verified: %s", file.DisplayName)
		return nil
	}

	return fmt.Errorf("no valid download URLs available, last error: %v", lastErr)
}
