// backend/internal/androidhost/logcat.go
package androidhost

import "strings"

// Android log priorities (android/log.h).
const (
	logDebug = 3
	logInfo  = 4
	logWarn  = 5
	logError = 6
)

// priority reads the level a console log line carries; any other output is info.
func priority(line string) int {
	fields := strings.Fields(line)
	for _, f := range fields[:min(len(fields), 4)] {
		switch f {
		case "DEBUG":
			return logDebug
		case "INFO":
			return logInfo
		case "WARN":
			return logWarn
		case "ERROR":
			return logError
		}
	}
	return logInfo
}
