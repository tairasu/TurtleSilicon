package epochsilicon

import "time"

// RequiredFile represents a file required for EpochSilicon
type RequiredFile struct {
	RelativePath string // Path relative to game directory
	DownloadURL  string
	DisplayName  string
	Hash         string            // MD5 hash for verification
	Size         int64             // File size in bytes
	URLs         map[string]string // Multiple CDN URLs
}

// FileMetadata represents metadata for checking file updates
type FileMetadata struct {
	Size         int64
	LastModified time.Time
}

// EpochManifest represents the API response structure
type EpochManifest struct {
	Version   string      `json:"Version"`
	Uid       string      `json:"Uid"`
	Files     []EpochFile `json:"Files"`
	CheckedAt string      `json:"checked_at"`
}

// EpochFile represents a file in the manifest
type EpochFile struct {
	Path   string            `json:"Path"`
	Hash   string            `json:"Hash"`
	Size   int64             `json:"Size"`
	Custom bool              `json:"Custom"`
	URLs   map[string]string `json:"Urls"`
}

// fileUpdateResult represents the result of checking a single file for updates
type fileUpdateResult struct {
	file        RequiredFile
	needsUpdate bool
	err         error
}

const (
	ManifestAPIURL = "https://updater.project-epoch.net/api/v2/manifest?environment=production"
	CDNPriority    = "cloudflare,digitalocean,none" // Preferred CDN order
)
