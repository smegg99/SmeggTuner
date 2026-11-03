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
defaultTargets: ["web"]
namespaces: {
	common: {targets: ["backend", "web"]}
	log: {targets: ["backend", "web"]}
	report: {targets: ["backend", "web"]}
}
scan: ["../frontend/app", "../backend/internal", "../backend/services", "../backend/core", "../backend/common/logger"]
// Include keys passed through error objects, menu data and conditional expressions.
keyPatterns: [
	#"'((?:common|tuner|session|record|settings|log|report|shortcut)\.[A-Za-z0-9_.]+)'"#,
	#"i18n\.T[fn]?\(\s*\"([A-Za-z0-9_.]+)\""#,
	#"\"((?:session|record|report|tuner)\.error\.[A-Za-z0-9_]+)\""#,
	#"\bKey:\s*\"([A-Za-z0-9_.]+)\""#,
]
// These families are assembled from runtime enums. Report templates are checked separately.
dynamicKeys: [
	"common.title.room.",
	"common.unitSymbol.",
	"settings.calibration.screen.bar.",
	"settings.calibration.screen.stat.",
	"settings.calibration.screen.steps.",
	"settings.calibration.status.",
	"settings.detection.",
	"settings.scaleNamings.",
	"shortcut.key.minus",
	"shortcut.key.plus",
	"tuner.boxes.hzOnly",
	"tuner.boxes.notHeard",
	"tuner.calibrate.status.",
	"tuner.panel.modeHints.",
	"tuner.panel.modes.",
	"tuner.side.",
	"tuner.soundcard.uncalibrated",
	"tuner.soundcard.unreliable",
	"tuner.state.",
	"tuner.unit.",
	"tuner.web.mic.",
	"tuner.web.processing.",
]
