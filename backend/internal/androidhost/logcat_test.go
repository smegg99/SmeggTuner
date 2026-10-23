// backend/internal/androidhost/logcat_test.go
package androidhost

import "testing"

func TestPriorityFollowsTheLoggerLevel(t *testing.T) {
	for line, want := range map[string]int{
		"2026-10-27 22:48:22 [smeggtuner] ERROR Could not load the configuration": logError,
		"2026-10-27 22:48:22 [smeggtuner] WARN  slow start":                       logWarn,
		"2026-10-27 22:48:22 [smeggtuner] DEBUG loading config":                   logDebug,
		"2026-10-27 22:48:22 [smeggtuner] INFO  starting":                         logInfo,
		"open Android cache: no such file":                                        logInfo,
		"":                                                                        logInfo,
	} {
		if got := priority(line); got != want {
			t.Errorf("priority(%q) = %d, want %d", line, got, want)
		}
	}
}
