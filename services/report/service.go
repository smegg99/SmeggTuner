// Package report writes reports of the active session; core/report owns print decisions and knows nothing of Wails.
package report

import (
	"smegg.me/smeggtuner/common/logger"
	corereport "smegg.me/smeggtuner/core/report"
	sessionsvc "smegg.me/smeggtuner/services/session"
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
		logger.Warn(logger.MsgReportRejected, logger.Str("format", opts.Format))
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
		logger.Warn(logger.MsgReportRejected, logger.Err(err))
		return nil, buildError(err)
	}

	path, err := saveDialog(defaultName(rep, opts.Format), opts.Format)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return &ResultDTO{Format: opts.Format}, nil // cancelled
	}
	path = withExtension(path, opts.Format)

	if err := write(path, rep, opts.Format); err != nil {
		logger.Error(logger.MsgReportFailed, logger.Str("path", path), logger.Err(err))
		return nil, err
	}
	logger.Info(logger.MsgReportWritten,
		logger.Str("path", path), logger.Str("format", opts.Format),
		logger.Int("notes", rep.Summary.Notes),
		logger.Int("merged", rep.Summary.Merged))

	out := &ResultDTO{Path: path, Format: opts.Format}
	if opts.Format != FormatHTML {
		return out, nil
	}
	// On an open failure the path still travels back, so the file stays findable.
	opened, err := openInBrowser(path)
	if err != nil {
		logger.Warn(logger.MsgReportOpenFailed, logger.Str("path", path), logger.Err(err))
		return out, ErrOpenFailed
	}
	out.Opened = opened
	if opened {
		logger.Debug(logger.MsgReportOpened, logger.Str("path", path))
	}
	return out, nil
}
