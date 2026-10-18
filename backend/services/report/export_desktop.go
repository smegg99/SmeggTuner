//go:build !android

// backend/services/report/export_desktop.go
package report

import corereport "smegg.me/smeggtuner/backend/core/report"

func exportReport(rep *corereport.Report, format string) (*ResultDTO, error) {
	return exportFile(rep, format)
}
