//go:build !android

// backend/services/report/pdf_desktop.go
package report

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	"github.com/chromedp/chromedp"

	"smegg.me/smeggtuner/backend/common/logger"
)

// printTimeout bounds the browser run; the page fetches nothing, so a slow run means Chrome will not start, not a big document.
const printTimeout = 15 * time.Second

// pdf loads the HTML from a temp file over file:// rather than SetContent, because a megabyte-scale logo data URI breaks the debugging protocol.
func pdf(html []byte, landscape bool) ([]byte, error) {
	dir, err := os.MkdirTemp("", "smeggtuner-pdf-")
	if err != nil {
		return nil, ErrWriteFailed
	}
	defer func() { _ = os.RemoveAll(dir) }()

	src := filepath.Join(dir, "card.html")
	if err := os.WriteFile(src, html, filePerm); err != nil {
		return nil, ErrWriteFailed
	}

	// Throwaway profile dir: else Chrome uses the user's profile and won't start while their browser is open.
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(filepath.Join(dir, "profile")),
		chromedp.Flag("disable-extensions", true),
	)

	alloc, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(alloc)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, printTimeout)
	defer cancelTimeout()

	if err := chromedp.Do(ctx, chromedp.Navigate("file://"+src), chromedp.WaitReady("body")); err != nil {
		return nil, failed(err)
	}
	// The sheet's own @page rule wins; A4 is its fallback. Margins are in the stylesheet.
	layout := []chromedp.PDFOption{chromedp.PDFPrintBackground(), chromedp.PDFPreferCSSPageSize(),
		chromedp.PDFPaper(chromedp.PaperA4), chromedp.PDFMargin(0)}
	if landscape {
		layout = append(layout, chromedp.PDFLandscape())
	}
	out, err := chromedp.Run(ctx, chromedp.PrintToPDF(layout...))
	if err != nil {
		return nil, failed(err)
	}
	if len(out) == 0 {
		return nil, ErrRenderFailed
	}
	return out, nil
}

// failed logs a browser run's error and tells a missing browser (user-fixable) from a run that failed.
func failed(err error) error {
	logger.Error(i18n.KeyLogReportFailed, logger.Err(err))
	if isMissingBrowser(err) {
		return ErrNoBrowser
	}
	return ErrRenderFailed
}

// isMissingBrowser detects no-Chrome: chromedp wraps an *exec.Error on launch failure.
func isMissingBrowser(err error) bool {
	var e *exec.Error
	return errors.As(err, &e)
}
