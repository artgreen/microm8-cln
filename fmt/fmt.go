package fmt

import (
	"fmt"

	"paleotronic.com/core/settings"
)

func Print(v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Print(v...)
}

func Println(v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Println(v...)
}

func RPrintln(v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Println(v...)
}

func RPrintf(format string, v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Printf(format, v...)
}

func Printf(format string, v ...interface{}) {
	if !settings.Verbose {
		return
	}
	fmt.Printf(format, v...)
}

func Sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}

// Errorf is a thin pass-through to fmt.Errorf so callers using the
// paleotronic.com/fmt wrapper can format wrapped errors without a second
// import of stdlib `fmt`. Supports the %w verb for error chaining.
func Errorf(format string, v ...interface{}) error {
	return fmt.Errorf(format, v...)
}
