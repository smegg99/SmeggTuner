package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Format is how this build reached the user's machine, which decides where its data may live.
type Format string

const (
	FormatBinary   Format = "binary"   // plain binary, portable / development
	FormatAppImage Format = "appimage" // Linux AppImage (read-only squashfs)
	FormatFlatpak  Format = "flatpak"  // Linux Flatpak sandbox
	FormatSnap     Format = "snap"     // Linux Snap package
	FormatDeb      Format = "deb"      // Debian .deb package
	FormatRPM      Format = "rpm"      // RPM package
	FormatAUR      Format = "aur"      // Arch Linux package
	FormatNSIS     Format = "nsis"     // Windows NSIS installer
	FormatExe      Format = "exe"      // Windows portable executable
	FormatMacApp   Format = "macapp"   // macOS .app bundle
	FormatSystem   Format = "system"   // Linux system location (/usr); deb, rpm and the AUR are indistinguishable at runtime
)

const appName = "smeggtuner"

// buildFormat is stamped at build time via -ldflags; empty on a plain go build, then detected from the environment.
var buildFormat string

// DetectFormat returns the active build format: the stamped compile-time value, else environment detection.
func DetectFormat() Format {
	if buildFormat != "" {
		return Format(buildFormat)
	}
	return detectFromEnv()
}

func detectFromEnv() Format {
	if os.Getenv("APPIMAGE") != "" {
		return FormatAppImage
	}
	if os.Getenv("FLATPAK_ID") != "" {
		return FormatFlatpak
	}
	if os.Getenv("SNAP") != "" {
		return FormatSnap
	}
	return detectFromPath()
}

// detectFromPath recognizes system-managed install locations by where the binary sits. This
// cannot be a compile-time stamp: NSIS installs the same exe that ships portable, and one
// Linux build feeds deb, rpm and the AUR alike. An unrecognized location is a portable run.
func detectFromPath() Format {
	exe, err := os.Executable()
	if err != nil {
		return FormatBinary
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return formatForLocation(runtime.GOOS, exe, os.Getenv)
}

// formatForLocation decides by path prefix alone (host filepath semantics would make it
// untestable for the other OS), so windows paths are case-folded and slash-normalized first.
func formatForLocation(goos, exe string, env func(string) string) Format {
	switch goos {
	case "windows":
		norm := func(p string) string { return strings.ToLower(strings.ReplaceAll(p, `\`, "/")) }
		for _, v := range []string{"ProgramW6432", "ProgramFiles", "ProgramFiles(x86)"} {
			if root := env(v); root != "" && pathUnder(norm(exe), norm(root)) {
				return FormatNSIS
			}
		}
	case "linux":
		if pathUnder(exe, "/usr") {
			return FormatSystem
		}
	}
	return FormatBinary
}

// pathUnder reports whether path sits below root on a path-segment boundary; both slash-separated.
func pathUnder(path, root string) bool {
	return strings.HasPrefix(path, strings.TrimSuffix(root, "/")+"/")
}

// IsInstalled reports whether the format lives in a read-only or system-managed location that cannot hold runtime data.
func (f Format) IsInstalled() bool {
	switch f {
	case FormatAppImage, FormatFlatpak, FormatSnap, FormatMacApp,
		FormatDeb, FormatRPM, FormatAUR, FormatNSIS, FormatSystem:
		return true
	default:
		return false
	}
}

// userConfigDir returns the platform user config directory with the app name appended.
func userConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName), nil
}

// userDataDir returns the platform user data directory with the app name appended (Linux XDG_DATA_HOME, Windows LocalAppData).
func userDataDir() (string, error) {
	switch runtime.GOOS {
	case "linux":
		if d := os.Getenv("XDG_DATA_HOME"); d != "" {
			return filepath.Join(d, appName), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", appName), nil
	case "windows":
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			return filepath.Join(d, appName), nil
		}
		return userConfigDir()
	default:
		return userConfigDir()
	}
}
