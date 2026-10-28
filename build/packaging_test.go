// build/packaging_test.go
package build_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/goreleaser/nfpm/v2"
	task "github.com/wailsapp/task/v3"
	"github.com/wailsapp/task/v3/taskfile/ast"
	"gopkg.in/yaml.v3"
)

func TestMain(m *testing.M) {
	if os.Getenv("SMEGGTUNER_PACKAGE_STUB") == "1" {
		lines := append([]string{os.Getenv("PACKAGE_ARCH"), os.Getenv("GOARCH")}, os.Args[1:]...)
		if err := os.WriteFile("observed", []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// Execute the maintained task's commands/vars/env with fake go, without its build dependencies.
func runPackagingTask(t *testing.T, platform, name string, variables map[string]any) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(platform, "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	definition := document["tasks"].(map[string]any)[name].(map[string]any)
	delete(definition, "deps")
	delete(definition, "preconditions")
	dir := t.TempDir()
	vars := map[string]any{"ROOT_DIR": dir, "BIN_DIR": "bin", "APP_NAME": "smeggtuner"}
	for key, value := range variables {
		vars[key] = value
	}
	fixture := map[string]any{"version": "3", "vars": vars, "tasks": map[string]any{"probe": definition}}
	encoded, err := yaml.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Taskfile.yml"), encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	// Reuse this test executable as a portable stub; no Go tool or packager is launched.
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	nameOnDisk := "go"
	if runtime.GOOS == "windows" {
		nameOnDisk += ".exe"
	}
	stub := filepath.Join(dir, nameOnDisk)
	if err := os.Link(executable, stub); err != nil {
		input, err := os.Open(executable)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = input.Close() }()
		output, err := os.OpenFile(stub, os.O_CREATE|os.O_WRONLY, 0o700)
		if err != nil {
			t.Fatal(err)
		}
		_, copyErr := io.Copy(output, input)
		closeErr := output.Close()
		if copyErr != nil || closeErr != nil {
			t.Fatalf("copy stub: %v; close: %v", copyErr, closeErr)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("SMEGGTUNER_PACKAGE_STUB", "1")
	var output bytes.Buffer
	e := task.Executor{Dir: dir, Stdout: &output, Stderr: &output, DisableVersionCheck: true}
	if err := e.Setup(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := e.Run(ctx, &ast.Call{Task: "probe"}); err != nil {
		t.Fatalf("task failed: %v\n%s", err, &output)
	}
	observed, err := os.ReadFile(filepath.Join(dir, "observed"))
	if err != nil {
		t.Fatalf("stub not called: %v\n%s", err, &output)
	}
	return strings.Split(strings.TrimSuffix(string(observed), "\n"), "\n")
}

func TestLinuxPackageTasksPassTargetArchitectureToNFPM(t *testing.T) {
	data, err := os.ReadFile("linux/nfpm/nfpm.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"deb", "rpm", "aur"} {
		for _, arch := range []string{"amd64", "arm64"} {
			t.Run(format+"/"+arch, func(t *testing.T) {
				// Deliberately disagree with the target to catch inherited host architecture.
				t.Setenv("GOARCH", "386")
				t.Setenv("PACKAGE_ARCH", "386")
				observed := runPackagingTask(t, "linux", "generate:"+format, map[string]any{"ARCH": arch})
				if observed[1] != runtime.GOARCH {
					t.Fatalf("packager must run on host, got GOARCH=%s", observed[1])
				}
				cfg, err := nfpm.ParseWithEnvMapping(bytes.NewReader(data), func(key string) string {
					if key == "PACKAGE_ARCH" {
						return observed[0]
					}
					return "1.4.0"
				})
				if err != nil || cfg.Arch != arch {
					t.Fatalf("package arch = %s, want %s; error = %v; args = %v", cfg.Arch, arch, err, observed)
				}
			})
		}
	}
}

func TestWindowsMSIXSelectsRequestedToolAndCurrentConfig(t *testing.T) {
	for _, useTool := range []string{"false", "true"} {
		t.Run(useTool, func(t *testing.T) {
			args := runPackagingTask(t, "windows", "create:msix:package", map[string]any{
				"ARCH": "arm64", "USE_MSIX_TOOL": useTool,
			})
			want := "--use-makeappx"
			if useTool == "true" {
				want = "--use-msix-tool"
			}
			if args[len(args)-1] != want {
				t.Fatalf("wrong tool: %v", args)
			}
			for i, arg := range args {
				if arg == "--config" && !strings.HasSuffix(args[i+1], "/build/config.yml") {
					t.Fatalf("obsolete config: %v", args)
				}
			}
		})
	}
}

func TestWindowsInstallerSigningUsesNSISOutput(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			args := runPackagingTask(t, "windows", "sign:installer", map[string]any{"ARCH": arch, "SIGN_CERTIFICATE": "test.pfx"})
			for i, arg := range args {
				if arg == "--input" {
					if want := "bin/smeggtuner-" + arch + "-installer.exe"; args[i+1] != want {
						t.Fatalf("signing %q instead of %q", args[i+1], want)
					}
					return
				}
			}
			t.Fatalf("missing signing input: %v", args)
		})
	}
}
