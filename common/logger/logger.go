package logger

import (
	"os"
	"strings"
	"sync"

	"github.com/smegg99/s99logger"
	"github.com/smegg99/s99logger/rotation"
)

const defaultService = "smeggtuner"

var (
	mu       sync.RWMutex
	global   = newConsoleLogger(defaultService, false, s99logger.LevelDebug)
	fileSink *rotation.Sink
)

// Options configures the global logger from application config.
type Options struct {
	Verbose     bool
	NoColor     bool
	Level       string
	Prefix      string
	EnableFiles bool
	Dir         string
	LogName     string
	MaxSizeMb   int64
	MaxBackups  int64
	MaxAgeDays  int64
	Compression string
	LocalTime   bool
}

func newConsoleLogger(service string, noColor bool, level s99logger.Level) *s99logger.Logger {
	console := s99logger.NewConsoleSink(os.Stderr)
	if noColor {
		console.WithColor(false)
	}
	return s99logger.New(console, s99logger.Options{Service: service, MinLevel: level})
}

func parseLevel(level string, verbose bool) s99logger.Level {
	if verbose {
		return s99logger.LevelDebug
	}
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "TRACE", "DEBUG":
		return s99logger.LevelDebug
	case "INFO", "":
		return s99logger.LevelInfo
	case "WARN", "WARNING":
		return s99logger.LevelWarn
	case "ERROR", "FATAL", "PANIC":
		return s99logger.LevelError
	default:
		return s99logger.LevelInfo
	}
}

func read() *s99logger.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return global
}

// Initialize installs a default console logger; safe to call before config is available.
func Initialize() {
	logger := newConsoleLogger(defaultService, false, s99logger.LevelDebug)
	mu.Lock()
	global = logger
	mu.Unlock()
}

// Configure rebuilds the global logger from opts, attaching a rotating file sink when enabled and closing any previous one.
func Configure(opts Options) error {
	console := s99logger.NewConsoleSink(os.Stderr)
	if opts.NoColor {
		console.WithColor(false)
	}

	sinks := []s99logger.Sink{console}

	compression := opts.Compression
	if compression == "" {
		compression = "zstd"
	}

	var newFile *rotation.Sink
	if opts.EnableFiles && opts.Dir != "" {
		fs, err := rotation.New(rotation.Options{
			Directory:   opts.Dir,
			Filename:    opts.LogName,
			MaxSizeMB:   int(opts.MaxSizeMb),
			MaxBackups:  int(opts.MaxBackups),
			MaxAgeDays:  int(opts.MaxAgeDays),
			Compression: compression,
			LocalTime:   opts.LocalTime,
		})
		if err != nil {
			return err
		}
		newFile = fs
		sinks = append(sinks, fs)
	}

	service := defaultService
	if opts.Prefix != "" {
		service = opts.Prefix
	}

	logger := s99logger.New(s99logger.MultiSink(sinks...), s99logger.Options{
		Service:  service,
		MinLevel: parseLevel(opts.Level, opts.Verbose),
	})

	mu.Lock()
	old := fileSink
	global = logger
	fileSink = newFile
	mu.Unlock()

	if old != nil {
		_ = old.Close()
	}
	return nil
}

// Underlying returns the current global logger for bridging into *s99logger.Logger consumers; call after Configure to pick up its sinks.
func Underlying() *s99logger.Logger {
	return read()
}

// Close releases the file sink, if any. Call during shutdown.
func Close() error {
	mu.Lock()
	fs := fileSink
	fileSink = nil
	mu.Unlock()
	if fs != nil {
		return fs.Close()
	}
	return nil
}
