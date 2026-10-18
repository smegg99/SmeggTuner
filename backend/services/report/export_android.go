//go:build android

// backend/services/report/export_android.go
package report

import (
	"bytes"
	"context"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/common/logger"
	corereport "smegg.me/smeggtuner/backend/core/report"
	"smegg.me/smeggtuner/backend/internal/androidhost"
)

func exportReport(rep *corereport.Report, format string) (*ResultDTO, error) {
	if format != FormatPDF {
		return exportFile(rep, format)
	}
	var html bytes.Buffer
	if err := corereport.HTML(&html, rep); err != nil {
		return nil, ErrRenderFailed
	}
	app := application.Get()
	if app == nil {
		return nil, ErrRenderFailed
	}
	ctx, cancel := context.WithTimeout(app.Context(), time.Minute)
	defer cancel()
	if err := androidhost.PrintDocument(ctx, html.Bytes(), rep.Layout.Landscape, defaultName(rep, format)); err != nil {
		logger.Error(i18n.KeyLogReportFailed, logger.Err(err))
		return nil, ErrRenderFailed
	}
	return &ResultDTO{Format: format, Printing: true}, nil
}
