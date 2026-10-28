// deploy/app/main.go

// The host for app.smeggtuner.com: it serves the web edition that fetch.sh unpacked, behind the box's proxy.
// Its own module, so the server builds it without the desktop app's sibling checkouts.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	addr, dir := os.Getenv("LISTEN_ADDR"), os.Getenv("SITE_DIR")
	if addr == "" || dir == "" {
		slog.Error("LISTEN_ADDR and SITE_DIR must be set")
		os.Exit(2)
	}
	site, err := newSite(dir)
	if err != nil {
		slog.Error("load site", "dir", dir, "err", err)
		os.Exit(1)
	}
	server := &http.Server{Addr: addr, Handler: site, ReadHeaderTimeout: 10 * time.Second}
	slog.Info("serving", "addr", addr, "dir", dir)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
}
