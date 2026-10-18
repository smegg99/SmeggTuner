//go:build android

// backend/services/report/pdf_android.go
package report

// Android PDF exports use the system print UI through exportReport.
func pdf(html []byte, landscape bool) ([]byte, error) {
	return nil, ErrInvalidFormat
}
