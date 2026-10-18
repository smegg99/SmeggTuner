// Package report writes reports of the active session; core/report owns print decisions and knows nothing of Wails.
package report

import (
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	corereport "smegg.me/smeggtuner/backend/core/report"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
)

// Native dialog strings; ASCII English until common/i18n carries keys for them.
const (
	saveTitle     = "Save the tuning report"
	htmlFilter    = "HTML report"
	htmlGlob      = "*.html"
	pdfFilter     = "PDF report"
	pdfGlob       = "*.pdf"
	csvFilter     = "CSV table"
	csvGlob       = "*.csv"
	dateLayout    = "2006-01-02"
	fileDatestamp = "2006-01-02"
	filePerm      = 0o644
)

// Service writes reports of the active session.
type Service struct {
	sessions *sessionsvc.Service
}

func New(sessions *sessionsvc.Service) *Service { return &Service{sessions: sessions} }

// Export writes a pass of the active session to a file the user chooses; for HTML it also opens it in the browser.
func (s *Service) Export(opts OptionsDTO) (*ResultDTO, error) {
	if opts.Format != FormatHTML && opts.Format != FormatPDF && opts.Format != FormatCSV {
		logger.Warn(i18n.KeyLogReportRejected, logger.Str("format", opts.Format))
		return nil, ErrInvalidFormat
	}

	snap := s.sessions.Snapshot()
	if snap == nil {
		return nil, sessionsvc.ErrNoSession
	}

	options, err := s.options(opts)
	if err != nil {
		return nil, err
	}
	rep, err := corereport.Build(snap, options)
	if err != nil {
		logger.Warn(i18n.KeyLogReportRejected, logger.Err(err))
		return nil, buildError(err)
	}
	return exportReport(rep, opts.Format)
}

func exportFile(rep *corereport.Report, format string) (*ResultDTO, error) {
	path, err := saveDialog(defaultName(rep, format), format)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return &ResultDTO{Format: format}, nil // cancelled
	}
	path = withExtension(path, format)

	if err := write(path, rep, format); err != nil {
		logger.Error(i18n.KeyLogReportFailed, logger.Str("path", path), logger.Err(err))
		return nil, err
	}
	logger.Info(i18n.KeyLogReportWritten,
		logger.Str("path", path), logger.Str("format", format),
		logger.Int("notes", rep.Summary.Notes))

	out := &ResultDTO{Path: path, Format: format}
	if format != FormatHTML {
		return out, nil
	}
	// On an open failure the path still travels back, so the file stays findable.
	opened, err := openInBrowser(path)
	if err != nil {
		logger.Warn(i18n.KeyLogReportOpenFailed, logger.Str("path", path), logger.Err(err))
		return out, ErrOpenFailed
	}
	out.Opened = opened
	if opened {
		logger.Debug(i18n.KeyLogReportOpened, logger.Str("path", path))
	}
	return out, nil
}
