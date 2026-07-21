package config

import (
	"os"
	"path/filepath"
	"runtime"
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
	return FormatBinary
}

// IsInstalled reports whether the format lives in a read-only or system-managed location that cannot hold runtime data.
func (f Format) IsInstalled() bool {
	switch f {
	case FormatAppImage, FormatFlatpak, FormatSnap, FormatMacApp,
		FormatDeb, FormatRPM, FormatAUR, FormatNSIS:
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
