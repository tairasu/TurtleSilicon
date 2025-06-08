//go:build !release

package debug

import "log"

// Printf logs with fmt.Printf style formatting in debug builds
func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Println logs with fmt.Println style in debug builds
func Println(v ...interface{}) {
	log.Println(v...)
}
