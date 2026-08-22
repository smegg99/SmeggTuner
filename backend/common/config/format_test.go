package config

import (
	"os"
	"path/filepath"
	"testing"
)

// The same binary ships portable and packaged, so install detection is a runtime
// path decision. These are the locations that decide it.
func TestFormatForLocation(t *testing.T) {
	winEnv := func(k string) string {
		switch k {
		case "ProgramW6432":
			return `C:\Program Files`
		case "ProgramFiles":
			return `C:\Program Files`
		case "ProgramFiles(x86)":
			return `C:\Program Files (x86)`
		}
		return ""
	}
	noEnv := func(string) string { return "" }

	homeEnv := func(k string) string {
		if k == "HOME" {
			return "/home/dev"
		}
		return ""
	}
	cases := []struct {
		name string
		goos string
		exe  string
		env  func(string) string
		want Format
	}{
		{"windows installed", "windows", `C:\Program Files\Smegg99\SmeggTuner\smeggtuner.exe`, winEnv, FormatNSIS},
		{"windows installed x86", "windows", `C:\Program Files (x86)\Smegg99\SmeggTuner\smeggtuner.exe`, winEnv, FormatNSIS},
		{"windows installed case-insensitive", "windows", `c:\program files\smegg99\smeggtuner\smeggtuner.exe`, winEnv, FormatNSIS},
		{"windows portable", "windows", `C:\Users\dev\Desktop\smeggtuner.exe`, winEnv, FormatBinary},
		{"windows prefix is not a boundary", "windows", `C:\Program Filesystem\smeggtuner.exe`, winEnv, FormatBinary},
		{"linux packaged", "linux", "/usr/bin/smeggtuner", noEnv, FormatSystem},
		{"linux local install", "linux", "/usr/local/bin/smeggtuner", noEnv, FormatSystem},
		{"linux portable", "linux", "/home/dev/bin/smeggtuner", noEnv, FormatBinary},
		{"linux dev tree", "linux", "/home/dev/Documents/Repositories/SmeggTuner/bin/smeggtuner", noEnv, FormatBinary},
		{"linux user install", "linux", "/home/dev/.local/lib/smeggtuner/smeggtuner", homeEnv, FormatUser},
		{"linux user install needs HOME", "linux", "/home/dev/.local/lib/smeggtuner/smeggtuner", noEnv, FormatBinary},
		{"linux other user lib", "linux", "/home/dev/.local/lib/other/smeggtuner", homeEnv, FormatBinary},
		{"mac bundle", "darwin", "/Applications/SmeggTuner.app/Contents/MacOS/smeggtuner", noEnv, FormatMacApp},
		{"mac custom bundle", "darwin", "/Users/dev/Tools/smeggtuner.app/Contents/MacOS/smeggtuner", noEnv, FormatMacApp},
		{"mac portable", "darwin", "/Users/dev/bin/smeggtuner", noEnv, FormatBinary},
		{"mac path boundary", "darwin", "/Tools/not.app/Contents/MacOSOther/smeggtuner", noEnv, FormatBinary},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatForLocation(c.goos, c.exe, c.env); got != c.want {
				t.Fatalf("formatForLocation(%s, %s) = %s, want %s", c.goos, c.exe, got, c.want)
			}
		})
	}
}

func TestCustomWindowsInstallUsesUserDataAndPortableDoesNot(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "smeggtuner.exe")
	noEnv := func(string) string { return "" }
	if got := formatForExecutable("windows", exe, noEnv); got != FormatBinary {
		t.Fatalf("unmarked portable executable = %s", got)
	}
	marker := filepath.Join(dir, "smeggtuner.install")
	for _, contents := range []string{"", "other", "nsis\r\n"} {
		if err := os.WriteFile(marker, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		got := formatForExecutable("windows", exe, noEnv)
		if contents == "nsis\r\n" {
			if got != FormatNSIS || !got.IsInstalled() {
				t.Fatalf("custom installer must select user config/data directories, got %s", got)
			}
		} else if got != FormatBinary {
			t.Fatalf("invalid marker changed portable mode: %s", got)
		}
	}
	if got := formatForExecutable("linux", exe, noEnv); got != FormatBinary {
		t.Fatalf("Windows marker affected Linux: %s", got)
	}
}
