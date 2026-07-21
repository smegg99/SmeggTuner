// assets/tray/trayicons.cue
//
// The tray icons, declared rather than scripted.
//
// There is ONE source SVG per theme and NO badge: both states draw the same
// mark. The light and dark sources exist for the only thing a tray icon can be
// sure of - a tray is somebody else's panel, whatever colour the desktop theme
// makes it, so the mark ships in the two inks that stay legible on either.
//
// The states remain because the engine still reports them and the generated
// Status constants are what the app switches on. They simply look alike for
// now. When a recording mark is wanted again, give the element in each source
// an inkscape:label, declare it in `slots`, and colour it per state - see the
// git history of this file for how that read.
prefix: "smeggtuner-tray"

// 64 is primary, and it is primary because the tray is handed a BITMAP.
//
// Wails publishes IconPixmap - one raster, which the panel then scales to
// whatever height it happens to be. A 32px icon on a scaled display is being
// enlarged to ~48px and looks it. Downscaling from 64 costs nothing and stays
// sharp. This whole line goes away if the tray is ever handed the SVG by name,
// which is what a StatusNotifierItem actually wants.
sizes: [64, 32]

slots: []

themes: {
	dark: source:  "smeggtuner-logo-tray-dark-badge.svg"
	light: source: "smeggtuner-logo-tray-light-badge.svg"
}

states: {
	idle: {}
	recording: {}
}
