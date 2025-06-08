//go:build release

package debug

// Printf is a no-op in release builds
func Printf(format string, v ...interface{}) {
	// No-op in release builds to reduce binary size
}

// Println is a no-op in release builds
func Println(v ...interface{}) {
	// No-op in release builds to reduce binary size
}
