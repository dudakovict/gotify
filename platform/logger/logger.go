// Package logger provides a convience function to constructing a logger
// for use. This is required not just for applications but for testing.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New constructs a Logger that writes to stdout and
// provides human readable timestamps.
func New() *zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	logger := zerolog.New(output).With().Timestamp().Logger()

	return &logger
}
