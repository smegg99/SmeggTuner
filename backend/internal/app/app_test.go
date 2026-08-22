// backend/internal/app/app_test.go
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/core/datastore"
	"smegg.me/smeggtuner/backend/core/repositories"
	coresession "smegg.me/smeggtuner/backend/core/session"
	sessionservice "smegg.me/smeggtuner/backend/services/session"
)

// A subprocess catches an accidental os.Exit without terminating the entire suite.
func TestStartupAndShutdownFailuresReturnAfterCleanup(t *testing.T) {
	const scenarioEnv = "SMEGGTUNER_LIFECYCLE_TEST"
	if scenario := os.Getenv(scenarioEnv); scenario != "" {
		checkLifecycle(t, scenario)
		return
	}
	for _, scenario := range []string{"config", "logger", "datastore", "run", "shutdown", "success"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStartupAndShutdownFailuresReturnAfterCleanup$", "-test.v")
			cmd.Env = append(os.Environ(), scenarioEnv+"="+scenario)
			output, err := cmd.CombinedOutput()
			if err != nil || !strings.Contains(string(output), "lifecycle returned after cleanup") {
				t.Fatalf("%s: %v\n%s", scenario, err, output)
			}
		})
	}
}

func checkLifecycle(t *testing.T, scenario string) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	configPath := filepath.Join(dir, "config.json")
	t.Setenv("CONFIG_PATH", configPath)
	dbPath := filepath.Join(dir, "sessions.db")
	blocked := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(blocked, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	settings := map[string]any{"storage": map[string]any{"db_path": dbPath}}
	if scenario == "datastore" {
		settings["storage"] = map[string]any{"db_path": filepath.Join(blocked, "sessions.db")}
	}
	if scenario == "logger" {
		settings["logger"] = map[string]any{"enable_files": true, "dir": filepath.Join(blocked, "logs")}
	}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if scenario == "config" {
		data = []byte("invalid JSON")
	}
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var opened *sql.DB
	var sessionID string
	var shown []startupFailure
	code := run(func() error {
		var err error
		opened, err = datastore.Get().DB()
		if err != nil {
			t.Fatal(err)
		}
		session, err := sessionSvc.Create(sessionservice.NewSessionDTO{
			Name: "Lifecycle test", Instrument: coresession.Instrument{ReedCount: 1, A4: 440},
		})
		if err != nil {
			t.Fatal(err)
		}
		sessionID = session.ID
		if err := sessionSvc.SetNotes("saved before database close"); err != nil {
			t.Fatal(err)
		}
		if scenario == "shutdown" {
			return opened.Close()
		}
		if scenario == "run" {
			return errors.New("injected application startup failure")
		}
		return nil
	}, func(f startupFailure) { shown = append(shown, f) })
	want := 1
	if scenario == "success" {
		want = 0
	}
	if code != want {
		t.Fatalf("exit code = %d, want %d", code, want)
	}
	// A failure before launch is shown, and only a bad settings file names that file.
	before := scenario == "config" || scenario == "logger" || scenario == "datastore"
	if before != (len(shown) == 1) || (len(shown) == 1 && (shown[0].path == configPath) != (scenario == "config")) {
		t.Fatalf("shown failures = %+v", shown)
	}
	if datastore.Path() != "" {
		t.Fatal("datastore was not closed")
	}
	if opened != nil {
		if err := opened.Ping(); err == nil {
			t.Fatal("SQL connection remained open")
		}
	} else if scenario == "run" || scenario == "shutdown" || scenario == "success" {
		t.Fatal("application callback was not reached")
	}
	if scenario == "run" || scenario == "success" {
		if err := datastore.Initialize(dbPath); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = datastore.Close() }()
		saved, err := repositories.GetSessionRepository().Get(sessionID)
		if err != nil || saved.Notes != "saved before database close" {
			t.Fatalf("saved session = %+v, error = %v", saved, err)
		}
	}
	t.Log("lifecycle returned after cleanup")
}

func TestTheFailurePageNamesTheFileAndEscapesTheError(t *testing.T) {
	page := string(renderFailure(startupFailure{path: "/home/u/config.json", err: errors.New("bad <key> tuner.note_sounds")}))
	for _, want := range []string{"/home/u/config.json", "bad &lt;key&gt; tuner.note_sounds", i18n.AppStartupFailedTitle(i18n.Localizer())} {
		if !strings.Contains(page, want) {
			t.Fatalf("page lacks %q:\n%s", want, page)
		}
	}
	if other := string(renderFailure(startupFailure{err: errors.New("disk full")})); strings.Contains(other, "config.json") || !strings.Contains(other, "disk full") {
		t.Fatalf("a failure without a settings file:\n%s", other)
	}
}
