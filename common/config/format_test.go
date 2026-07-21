package config

import "testing"

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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatForLocation(c.goos, c.exe, c.env); got != c.want {
				t.Fatalf("formatForLocation(%s, %s) = %s, want %s", c.goos, c.exe, got, c.want)
			}
		})
	}
}
