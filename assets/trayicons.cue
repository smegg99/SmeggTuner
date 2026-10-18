// assets/trayicons.cue
prefix:    "smeggtuner"
goPackage: "trayicons"
// Sizes and padding use traygen defaults; per-size SVG overrides (platforms.<os>.sizeSources) live here.
slots: ["body"]
themes: {
	dark: {
		source: "smeggtuner-cat.svg"
		colors: {for state, _ in states {"\(state)": body: "#E9EBEE"}}
	}
	light: {
		source: "smeggtuner-cat.svg"
		colors: {for state, _ in states {"\(state)": body: "#0F1218"}}
	}
}
states: idle: {}
// The same master is installed as the menu icon.
appIcon: source: "smeggtuner-app-icon.svg"
