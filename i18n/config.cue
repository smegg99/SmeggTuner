// i18n/config.cue
defaultLocale: "en"
locales: ["en", "pl"]
targets: ["web", "backend"]
roots: ["locales"]
output: "generated"
outputs: {
	web: "../frontend/i18n/generated"
	backend: "../backend/common/i18n"
}
go: {package: "i18n"}
namespaces: {
	log: {targets: ["backend", "web"]}
	app: {targets: ["backend"]}
	tray: {targets: ["backend"]}
	about: {targets: ["web"]}
	calibrate: {targets: ["web"]}
	common: {targets: ["web"]}
	curve: {targets: ["web"]}
	file: {targets: ["web"]}
	instrument: {targets: ["web"]}
	record: {targets: ["web"]}
	routes: {targets: ["web"]}
	session: {targets: ["web"]}
	sessionPage: {targets: ["web"]}
	settings: {targets: ["web"]}
	// Key cap names over the layer's own shortcut messages, which its UiKbd reads as shortcut.key.<name>.
	shortcut: {targets: ["web"]}
	title: {targets: ["web"]}
	tuner: {targets: ["web"]}
	web: {targets: ["web"]}
	workshop: {targets: ["web"]}
}
scan: ["../frontend/app", "../backend/internal", "../backend/services", "../backend/core", "../backend/common/logger"]
keyPatterns: [
	#"i18n\.T[fn]?\(\s*\"([A-Za-z0-9_.]+)\""#,
	#"log\.(?:error|warn|info|debug|trace)\(\s*['"](log\.[A-Za-z0-9_.]+)['"]"#,
]
