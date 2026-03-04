package model

// StatusLine represents a status line that can display messages.
type StatusLine interface {

	// Info sets an informational message on the status line, typically used to display non-critical updates or operations in progress.
	Info(msg string)
	// Infof formats a message and sets it on the status line as an informational message.
	Infof(format string, args ...any)
	// Warning sets a warning message on the status line, indicating potential issues or non-critical problems.
	Warning(msg string)
	// Warningf formats a message and sets it on the status line as a warning.
	Warningf(format string, args ...any)
	// Error sets an error message on the status line, indicating that an operation failed or encountered an issue.
	Error(msg string)
	// Errorf formats a message and sets it on the status line as an error.
	Errorf(format string, args ...any)
	// Clear clears the status line, removing all messages.
	Clear()
}
