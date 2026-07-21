package logger

import (
	"os"

	"github.com/smegg99/s99logger"
)

// Debug logs id at debug level with the given attributes.
func Debug(id s99logger.MessageID, attrs ...s99logger.Attr) {
	read().Debug(s99logger.NewEvent(id, attrs...))
}

// Info logs id at info level with the given attributes.
func Info(id s99logger.MessageID, attrs ...s99logger.Attr) {
	read().Info(s99logger.NewEvent(id, attrs...))
}

// Warn logs id at warn level with the given attributes.
func Warn(id s99logger.MessageID, attrs ...s99logger.Attr) {
	read().Warn(s99logger.NewEvent(id, attrs...))
}

// Error logs id at error level with the given attributes.
func Error(id s99logger.MessageID, attrs ...s99logger.Attr) {
	read().Error(s99logger.NewEvent(id, attrs...))
}

// Fatal logs id at error level and terminates the process.
func Fatal(id s99logger.MessageID, attrs ...s99logger.Attr) {
	read().Error(s99logger.NewEvent(id, attrs...))
	_ = Close()
	os.Exit(1)
}

// Attr is a re-exported structured log attribute.
type Attr = s99logger.Attr

// MessageID is a re-exported log event identifier.
type MessageID = s99logger.MessageID

// Attribute constructors, re-exported so callers depend only on this package.

func Str(key, value string) s99logger.Attr { return s99logger.String(key, value) }

func Int(key string, value int) s99logger.Attr { return s99logger.Int(key, value) }

func Int64(key string, value int64) s99logger.Attr { return s99logger.Any(key, value) }

func Bool(key string, value bool) s99logger.Attr { return s99logger.Bool(key, value) }

func Err(err error) s99logger.Attr { return s99logger.Err(err) }

func Any(key string, value any) s99logger.Attr { return s99logger.Any(key, value) }
