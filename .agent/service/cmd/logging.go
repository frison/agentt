package cmd

import (
	"io"
	"log"
	"os"
)

// Log Levels
const (
	LevelError = iota // 0: Default level, includes errors
	LevelWarn         // 1: Includes warnings and errors
	LevelInfo         // 2: Includes info, warnings, errors
	LevelDebug        // 3: Includes debug, info, warnings, errors
)

// currentLogLevel is set by the root command based on flags.
var currentLogLevel = LevelWarn // Default to warnings and errors

// configureLogging sets the log output and level based on flags.
// Called from root PersistentPreRunE.
func configureLogging(quiet bool, verbosity int) {
	if quiet {
		// Quiet mode: Suppress WARN, INFO, DEBUG. Only ERROR (and FATAL via os.Exit) remain.
		currentLogLevel = LevelError
		log.SetOutput(os.Stderr) // Ensure errors still go somewhere
	} else {
		log.SetOutput(os.Stderr) // Default output
		switch verbosity {
		case 0:
			currentLogLevel = LevelWarn // Default: Warn+
		case 1:
			currentLogLevel = LevelInfo // -v: Info+
		default: // >= 2
			currentLogLevel = LevelDebug // -vv and above: Debug+
		}
	}

	// Ensure log package uses standard flags (timestamp, etc.)
	// log.SetFlags(log.LstdFlags | log.Lshortfile) // Optional: Add file/line info for debug
	log.SetFlags(log.Ltime) // Just time for cleaner output
}

// --- Logging Helper Functions ---

func logDebug(format string, v ...interface{}) {
	if currentLogLevel >= LevelDebug {
		log.Printf("DEBUG: "+format, v...)
	}
}

func logInfo(format string, v ...interface{}) {
	if currentLogLevel >= LevelInfo {
		// No prefix for standard info messages
		log.Printf(format, v...)
	}
}

func logWarn(format string, v ...interface{}) {
	if currentLogLevel >= LevelWarn {
		log.Printf("WARN: "+format, v...)
	}
}

// Note: Errors leading to command failure are typically returned as `error`
// and handled by Cobra/main.go. This function is for non-fatal errors
// that should be logged according to level but don't stop execution.
// For fatal errors, prefer returning an error from RunE.
func logError(format string, v ...interface{}) {
	// Errors are always logged unless completely silenced elsewhere (which we avoid)
	if currentLogLevel >= LevelError { // Should always be true unless level is < 0
		log.Printf("ERROR: "+format, v...)
	}
}

// discardLogs can be used if truly needed, but configureLogging handles -q now.
func discardLogs() {
	log.SetOutput(io.Discard)
}
