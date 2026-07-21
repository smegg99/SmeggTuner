package report

import (
	audiosvc "smegg.me/smeggtuner/services/audio"
)

// ServiceError is the i18n-keyed error all services share (aliased from audio) so errors.As reaches every service's keys.
type ServiceError = audiosvc.ServiceError

var (
	// ErrInvalidFormat reports a format that is none of pdf, html or csv.
	ErrInvalidFormat = &ServiceError{Key: "report.error.invalidFormat"}
	// ErrInvalidDate reports a date that is not yyyy-mm-dd.
	ErrInvalidDate = &ServiceError{Key: "report.error.invalidDate"}
	// ErrLogoUnreadable reports a letterhead logo that would not open or is not an image.
	ErrLogoUnreadable = &ServiceError{Key: "report.error.logoUnreadable"}
	// ErrRenderFailed reports a report that would not render.
	ErrRenderFailed = &ServiceError{Key: "report.error.renderFailed"}
	// ErrWriteFailed reports a file that would not be written.
	ErrWriteFailed = &ServiceError{Key: "report.error.writeFailed"}
	// ErrNoBrowser reports no headless browser to render the PDF; kept distinct because it is user-fixable.
	ErrNoBrowser = &ServiceError{Key: "report.error.noBrowser"}
	// ErrOpenFailed reports a written report that no browser would open; the path still comes back with it.
	ErrOpenFailed = &ServiceError{Key: "report.error.openFailed"}
)

// Format is what the report is written as; FormatPDF is the HTML sheet laid out by a browser (pdf.go).
const (
	FormatHTML = "html"
	FormatCSV  = "csv"
	FormatPDF  = "pdf"
)

// OptionsDTO is what the report dialog emits, matching the frontend's ReportOptions.
type OptionsDTO struct {
	// Format is "pdf", "html" or "csv".
	Format string `json:"format"`
	// Date printed on the sheet, yyyy-mm-dd; empty means the pass's own date.
	Date string `json:"date"`
}

// ResultDTO is where the report went.
type ResultDTO struct {
	// Path is where it was written; empty means the user cancelled.
	Path   string `json:"path"`
	Format string `json:"format"`
	// Opened says the HTML sheet was handed to the browser; only the HTML is.
	Opened bool `json:"opened"`
}
