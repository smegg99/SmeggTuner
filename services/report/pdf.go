// PDF output reuses core/report's HTML printed by a headless Chrome, to stay in step with the browser.
package report

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"smegg.me/smeggtuner/common/logger"
)

// printTimeout bounds the browser run; the page fetches nothing, so a slow run means Chrome will not start, not a big document.
const printTimeout = 15 * time.Second

// A4 in inches, Chrome's print unit; landscape is chosen by the report layout, not here.
const (
	paperWidthIn  = 8.27
	paperHeightIn = 11.69
)

// pdf loads the HTML from a temp file over file:// rather than SetContent, because a megabyte-scale logo data URI breaks the debugging protocol.
func pdf(html []byte, landscape bool) ([]byte, error) {
	dir, err := os.MkdirTemp("", "smeggtuner-pdf-")
	if err != nil {
		return nil, ErrWriteFailed
	}
	defer os.RemoveAll(dir)

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

	var out []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("file://"+src),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			out, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				// Prefer the sheet's own @page rule; the explicit size is its fallback.
				WithPreferCSSPageSize(true).
				WithPaperWidth(paperWidthIn).
				WithPaperHeight(paperHeightIn).
				WithLandscape(landscape).
				// Margins are in the stylesheet; Chrome's defaults would add half an inch.
				WithMarginTop(0).WithMarginBottom(0).
				WithMarginLeft(0).WithMarginRight(0).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		// Distinguish a missing browser (user-fixable) from a run that failed.
		if isMissingBrowser(err) {
			logger.Error(logger.MsgReportFailed, logger.Err(err))
			return nil, ErrNoBrowser
		}
		logger.Error(logger.MsgReportFailed, logger.Err(err))
		return nil, ErrRenderFailed
	}
	if len(out) == 0 {
		return nil, ErrRenderFailed
	}
	return out, nil
}

// isMissingBrowser detects no-Chrome: chromedp wraps an *exec.Error on launch failure.
func isMissingBrowser(err error) bool {
	var e *exec.Error
	return errors.As(err, &e)
}
