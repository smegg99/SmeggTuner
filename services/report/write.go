package report

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"smegg.me/smeggtuner/common/logger"
	corereport "smegg.me/smeggtuner/core/report"
)

// write renders to a temp file beside the target then renames it over, so a failed write never clobbers the previous report.
func write(path string, rep *corereport.Report, format string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".report-*")
	if err != nil {
		return ErrWriteFailed
	}
	name := tmp.Name()
	defer os.Remove(name) // a no-op once the rename below has taken it away

	if err := renderTo(tmp, rep, format); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return ErrWriteFailed
	}
	if err := tmp.Close(); err != nil {
		return ErrWriteFailed
	}
	if err := os.Chmod(name, filePerm); err != nil {
		return ErrWriteFailed
	}
	if err := os.Rename(name, path); err != nil {
		return ErrWriteFailed
	}
	return nil
}

// renderTo writes the card; the PDF is the same HTML laid out by a browser (pdf.go).
func renderTo(w io.Writer, rep *corereport.Report, format string) error {
	if format == FormatCSV {
		if err := corereport.CSV(w, rep); err != nil {
			logger.Error(logger.MsgReportFailed, logger.Err(err))
			return ErrRenderFailed
		}
		return nil
	}

	var html bytes.Buffer
	if err := corereport.HTML(&html, rep); err != nil {
		logger.Error(logger.MsgReportFailed, logger.Err(err))
		return ErrRenderFailed
	}
	if format != FormatPDF {
		_, err := w.Write(html.Bytes())
		if err != nil {
			return ErrWriteFailed
		}
		return nil
	}

	pages, err := pdf(html.Bytes(), rep.Layout.Landscape)
	if err != nil {
		return err
	}
	if _, err := w.Write(pages); err != nil {
		return ErrWriteFailed
	}
	return nil
}
